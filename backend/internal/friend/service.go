package friend

import (
	"backend/internal/friend/agent/graph"
	"backend/internal/model"
	"backend/pkg/constants"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type FriendService interface {
	GetOrCreate(ctx context.Context, charID, userID uint) (*model.Friend, error)
	GetList(ctx context.Context, userID uint, cursorUpdatedAt *time.Time, cursorID uint) ([]*model.Friend, error)
	RemoveFriend(ctx context.Context, friendID, userID uint) error
	StreamChat(ctx context.Context, friendID, userID uint, userMsg string) (*schema.StreamReader[*schema.Message], string, error)
	SaveMessage(ctx context.Context, msg *model.Message) error
	UpdateMemory(ctx context.Context, friendID uint) error
	GetMessageCount(ctx context.Context, friendID uint) (int64, error)
	GetMessageHistory(ctx context.Context, friendID, userID, cursor uint, limit int) ([]*model.Message, error)
}

type friendService struct {
	repo        FriendRepository
	chatGraph   compose.Runnable[graph.ChatState, *schema.Message]
	memoryGraph compose.Runnable[graph.MemoryState, *schema.Message]
}

func NewFriendService(repo FriendRepository, chatGraph compose.Runnable[graph.ChatState, *schema.Message], memoryGraph compose.Runnable[graph.MemoryState, *schema.Message]) FriendService {
	return &friendService{
		repo:        repo,
		chatGraph:   chatGraph,
		memoryGraph: memoryGraph,
	}
}

func (s *friendService) GetOrCreate(ctx context.Context, charID, userID uint) (*model.Friend, error) {
	newFriend := &model.Friend{
		CharacterID: charID,
		MeID:        userID,
	}

	if err := s.repo.AddFriend(ctx, newFriend); err != nil {
		zap.L().Error("[friend service] AddFriend error", zap.Error(err))
		return nil, errors.New("系统繁忙，请稍后再试")
	}

	f, err := s.repo.GetFriend(ctx, charID, userID)
	if err != nil {
		zap.L().Error("[friend service] Final GetFriend error", zap.Error(err))
		return nil, errors.New("系统繁忙，请稍后再试")
	}

	return f, nil
}

func (s *friendService) GetList(ctx context.Context, userID uint, cursorUpdatedAt *time.Time, cursorID uint) ([]*model.Friend, error) {
	friends, err := s.repo.GetList(ctx, userID, cursorUpdatedAt, cursorID, constants.DefaultLimit)
	if err != nil {
		zap.L().Error("[friend service] GetList db error", zap.Error(err))
		return nil, errors.New("系统繁忙，请稍后再试")
	}
	return friends, nil
}

func (s *friendService) RemoveFriend(ctx context.Context, friendID, userID uint) error {
	friend, err := s.repo.GetByID(ctx, friendID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("好友不存在")
		}
		zap.L().Error("[friend service] GetByID db error", zap.Error(err))
		return errors.New("系统繁忙，请稍后再试")
	}

	if friend.MeID != userID {
		return errors.New("好友不存在")
	}

	if err := s.repo.RemoveFriend(ctx, friendID); err != nil {
		zap.L().Error("[friend service] DeleteFriend db error", zap.Error(err))
		return errors.New("系统繁忙，请稍后再试")
	}
	return nil
}

func (s *friendService) SaveMessage(ctx context.Context, msg *model.Message) error {
	if err := s.repo.SaveMessageTx(ctx, msg); err != nil {
		zap.L().Error("[friend service] SaveMessageTx error", zap.Error(err))
		return errors.New("系统繁忙，请稍后再试")
	}
	return nil
}

func (s *friendService) StreamChat(
	ctx context.Context,
	friendID, userID uint,
	userMsg string,
) (*schema.StreamReader[*schema.Message], string, error) {
	friend, err := s.repo.GetByID(ctx, friendID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, "", errors.New("好友不存在")
		}
		zap.L().Error("[friend service] GetByID db error", zap.Error(err))
		return nil, "", errors.New("系统繁忙，请稍后再试")
	}
	if friend.MeID != userID {
		zap.L().Error("[friend service] user 无权访问好友", zap.Uint("userID", userID))
		return nil, "", errors.New("好友不存在")
	}

	// 拼接 System Prompt
	prompts, err := s.repo.GetSystemPrompts(ctx, constants.SystemPromptTitleReply)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			zap.L().Error("[friend service] GetSystemPrompts no exists", zap.Error(err))
			return nil, "", errors.New("系统繁忙，请稍后再试")
		}
		zap.L().Error("[friend service] GetSystemPrompts db error", zap.Error(err))
		return nil, "", errors.New("系统繁忙，请稍后再试")
	}
	if len(prompts) == 0 {
		zap.L().Error("[friend service] reply system prompts is empty")
		return nil, "", errors.New("系统繁忙，请稍后再试")
	}

	var builder strings.Builder
	builder.Grow(len(prompts) * 100)
	for _, p := range prompts {
		builder.WriteString(p.Prompt)
	}
	basePrompt := builder.String()

	var name, profile string
	if friend.Character != nil {
		name, profile = friend.Character.Name, friend.Character.Profile
	}

	finalSystemPrompt := fmt.Sprintf("%s\n【角色名字】\n%s\n【角色性格】\n%s\n【长期记忆】\n%s\n",
		basePrompt, name, profile, friend.Memory)
	messages := []*schema.Message{
		schema.SystemMessage(finalSystemPrompt),
	}

	recentMsgs, err := s.repo.GetRecentMessages(ctx, friendID, constants.MaxChatHistoryCount)
	if err != nil {
		zap.L().Error("[friend service] GetRecentMessages db error", zap.Error(err), zap.Uint("friendID", friendID))
		return nil, "", errors.New("系统繁忙，请稍后再试")
	}

	currentLength := 0

	// 找到截断点：从后往前找，找到刚好不超过 maxContextLength 的起始索引
	startIndex := len(recentMsgs)
	for i := len(recentMsgs) - 1; i >= 0; i-- {
		m := recentMsgs[i]
		msgLen := len(m.UserMessage) + len(m.Output)
		if currentLength+msgLen > constants.MaxContextLength {
			break
		}
		currentLength += msgLen
		startIndex = i // 记录安全起始点
	}

	// 统一正序拼装，避免低效的头插法
	for i := startIndex; i < len(recentMsgs); i++ {
		m := recentMsgs[i]
		messages = append(messages,
			schema.UserMessage(m.UserMessage),
			schema.AssistantMessage(m.Output, nil),
		)
	}

	// 放入当前问题
	messages = append(messages, schema.UserMessage(userMsg))

	// 序列化 Input 给后续存数据库用
	inputBytes, err := json.Marshal(messages)
	if err != nil {
		zap.L().Error("[friend service] marshal input messages failed", zap.Error(err))
		return nil, "", errors.New("系统繁忙，请稍后再试")
	}
	inputStr := string(inputBytes)
	if len(inputStr) > constants.MaxDBInputLength { // 截断防爆
		inputStr = inputStr[:constants.MaxDBInputLength]
	}

	state := graph.ChatState{
		Messages: messages,
	}
	stream, err := s.chatGraph.Stream(ctx, state)
	if err != nil {
		zap.L().Error("[friend service] chat graph stream failed", zap.Error(err), zap.Uint("friendID", friendID))
		return nil, "", errors.New("系统繁忙，请稍后再试")
	}

	return stream, inputStr, err
}

func (s *friendService) UpdateMemory(ctx context.Context, friendID uint) error {
	friend, err := s.repo.GetByID(ctx, friendID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("好友不存在")
		}
		zap.L().Error("[friend service] GetByID db error in UpdateMemory", zap.Error(err), zap.Uint("friendID", friendID))
		return errors.New("系统繁忙，请稍后再试")
	}

	recentMsgs, err := s.repo.GetRecentMessages(ctx, friendID, constants.MaxMemorySummaryCount)
	if err != nil {
		zap.L().Error("[friend service] GetRecentMessages db error in UpdateMemory", zap.Error(err), zap.Uint("friendID", friendID))
		return errors.New("系统繁忙，请稍后再试")
	}
	if len(recentMsgs) == 0 {
		return nil
	}

	// "记忆"系统提示词
	prompts, err := s.repo.GetSystemPrompts(ctx, constants.SystemPromptTitleMemory)
	if err != nil {
		zap.L().Error("[friend service] GetSystemPrompts db error in UpdateMemory", zap.Error(err))
		return errors.New("系统繁忙，请稍后再试")
	}
	if len(prompts) == 0 {
		zap.L().Error("[friend service] memory system prompts is empty")
		return errors.New("系统繁忙，请稍后再试")
	}

	// 组装输入上下文
	messages := []*schema.Message{
		schema.SystemMessage(prompts[0].Prompt),
		schema.UserMessage(fmt.Sprintf("当前记忆JSON：\n%s\n\n请根据以下最新对话，按照规则输出更新后的 JSON：", friend.Memory)),
	}

	for _, m := range recentMsgs {
		messages = append(messages, schema.UserMessage(m.UserMessage))
		messages = append(messages, schema.AssistantMessage(m.Output, nil))
	}

	state := graph.MemoryState{Messages: messages}
	outMsg, err := s.memoryGraph.Invoke(ctx, state)
	if err != nil {
		zap.L().Error("[friend service] memory graph invoke failed", zap.Error(err), zap.Uint("friendID", friendID))
		return errors.New("系统繁忙，请稍后再试")
	}
	if outMsg == nil {
		zap.L().Error("[friend service] memory graph output is nil", zap.Uint("friendID", friendID))
		return errors.New("系统繁忙，请稍后再试")
	}

	// 清理大模型吐出的 Markdown 代码块符号 (比如 ```json ... ```)
	cleanJSON := strings.TrimPrefix(outMsg.Content, constants.MarkdownJSONPrefix)
	cleanJSON = strings.TrimPrefix(cleanJSON, constants.MarkdownPrefix)
	cleanJSON = strings.TrimSuffix(cleanJSON, constants.MarkdownSuffix)
	cleanJSON = strings.TrimSpace(cleanJSON)

	return s.repo.UpdateFriendMemory(ctx, friendID, cleanJSON)
}

func (s *friendService) GetMessageCount(ctx context.Context, friendID uint) (int64, error) {
	count, err := s.repo.GetMessageCount(ctx, friendID)
	if err != nil {
		zap.L().Error("[friend service] GetMessageCount db error",
			zap.Error(err),
			zap.Uint("friendID", friendID))
		return 0, errors.New("系统繁忙，请稍后再试")
	}
	return count, nil
}

func (s *friendService) GetMessageHistory(ctx context.Context, friendID, userID, cursor uint, limit int) ([]*model.Message, error) {
	friend, err := s.repo.GetByID(ctx, friendID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("好友不存在")
		}
		zap.L().Error("[friend service] GetByID db error in GetMessageHistory", zap.Error(err))
		return nil, errors.New("系统繁忙，请稍后再试")
	}
	if friend.MeID != userID {
		zap.L().Error("[friend service] user 无权访问好友历史消息", zap.Uint("userID", userID), zap.Uint("friendID", friendID))
		return nil, errors.New("好友不存在")
	}

	msgs, err := s.repo.GetMessageHistory(ctx, friendID, cursor, limit)
	if err != nil {
		zap.L().Error("[friend service] GetMessageHistory db error", zap.Error(err))
		return nil, errors.New("系统繁忙，请稍后再试")
	}

	return msgs, nil
}
