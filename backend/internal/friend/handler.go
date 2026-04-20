package friend

import (
	"backend/internal/model"
	"backend/pkg/constants"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type friendHandler struct {
	svc FriendService
}

func NewFriendHandler(svc FriendService) *friendHandler {
	return &friendHandler{svc: svc}
}

func (h *friendHandler) buildFriendResp(f *model.Friend) (FriendResp, bool) {
	if f == nil || f.Character == nil || f.Character.Author == nil {
		return FriendResp{}, false
	}
	char := f.Character
	author := char.Author
	return FriendResp{
		ID: f.ID,
		Character: CharacterResp{
			ID:      char.ID,
			Name:    char.Name,
			Profile: char.Profile,
			Photo:   constants.StaticBaseURL + char.Photo,
			BgImage: constants.StaticBaseURL + char.BackgroundImage,
			Author: AuthorResp{
				UserID:   author.ID,
				Username: author.Username,
				Photo:    constants.StaticBaseURL + author.Photo,
			},
		},
	}, true
}

func (h *friendHandler) GetOrCreate(c *gin.Context) {
	uid, _ := c.Get("user_id")
	userID, ok := uid.(uint)
	if !ok {
		zap.L().Error("[friend handler] userID type error")
		c.JSON(http.StatusOK, gin.H{"result": "系统繁忙，请稍后再试"})
		return
	}

	var req GetOrCreateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"result": "参数格式错误"})
		return
	}

	friend, err := h.svc.GetOrCreate(c.Request.Context(), req.CharacterID, userID)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"result": err.Error()})
		return
	}
	fmt.Println(friend)
	if friend != nil {
		fmt.Println(friend.Character)
		if friend.Character != nil {
			fmt.Println(friend.Character.Author)
		}
	}

	friendResp, ok := h.buildFriendResp(friend)
	if !ok {
		zap.L().Debug("[friend handler] buildFriendResp error")
		c.JSON(http.StatusOK, gin.H{"result": "系统繁忙，请稍后再试"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"result": "success",
		"friend": friendResp,
	})
}

func (h *friendHandler) GetFriendList(c *gin.Context) {
	uid, _ := c.Get("user_id")
	userID, ok := uid.(uint)
	if !ok {
		zap.L().Error("[friend handler] userID type error")
		c.JSON(http.StatusOK, gin.H{"result": "系统繁忙，请稍后再试"})
		return
	}

	itemsCountStr := c.DefaultQuery("items_count", "0")
	itemsCount, _ := strconv.Atoi(itemsCountStr)
	if itemsCount < 0 {
		itemsCount = 0
	}

	rawFriends, err := h.svc.GetList(c.Request.Context(), userID, int(itemsCount))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"result": err.Error()})
		return
	}

	friends := make([]FriendResp, 0, len(rawFriends))
	for _, f := range rawFriends {
		friendResp, ok := h.buildFriendResp(f)
		if !ok {
			continue
		}
		friends = append(friends, friendResp)
	}

	c.JSON(http.StatusOK, gin.H{
		"result":  "success",
		"friends": friends,
	})
}

func (h *friendHandler) RemoveFriend(c *gin.Context) {
	uid, _ := c.Get("user_id")
	userID, ok := uid.(uint)
	if !ok {
		zap.L().Error("[friend handler] userID type error")
		c.JSON(http.StatusOK, gin.H{"result": "系统繁忙，请稍后再试"})
		return
	}

	var req RemoveReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"result": "参数格式错误"})
		return
	}

	if err := h.svc.RemoveFriend(c.Request.Context(), req.FriendID, userID); err != nil {
		c.JSON(http.StatusOK, gin.H{"result": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"result": "success"})
}

func (h *friendHandler) StreamChat(c *gin.Context) {
	uid, _ := c.Get("user_id")
	userID, ok := uid.(uint)
	if !ok {
		zap.L().Error("[friend handler] userID type error in StreamChat")
		c.JSON(http.StatusOK, gin.H{"result": "系统繁忙，请稍后再试"})
		return
	}

	var req ChatReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"result": "参数格式错误"})
		return
	}

	// 获取流和输入
	stream, inputStr, err := h.svc.StreamChat(c.Request.Context(), req.FriendID, userID, req.Message)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"result": err.Error()})
		return
	}
	defer stream.Close() // 关流

	c.Writer.Header().Set("Content-Type", "text/event-stream") // 设置响应头为 SSE
	c.Writer.Header().Set("Cache-Control", "no-cache")         // 禁止缓存 -> nginx
	c.Writer.Header().Set("X-Accel-Buffering", "no")
	c.Writer.Header().Set("Connection", "keep-alive") // 保持长连接获取信息
	c.Writer.Flush()

	var finalOutput string
	var inputTokens, outputTokens, totalTokens int

	// 判断字符串转为 rune 的长度，根据 maxLen 截取
	truncateString := func(str string, maxLen int) string {
		runes := []rune(str)
		if len(runes) > maxLen {
			return string(runes[:maxLen])
		}
		return str
	}

	// 保存消息并检查是否更新记忆，每10条更新一次记忆
	saveAndUpdateMemory := func(isInterrupted bool, currentFinalOutput string, inTok, outTok, totTok int) {
		saveCtx := context.Background()

		msg := &model.Message{
			FriendID:     req.FriendID,
			UserMessage:  truncateString(req.Message, constants.MaxMsgLen),
			Input:        inputStr,
			Output:       truncateString(currentFinalOutput, constants.MaxMsgLen),
			InputTokens:  inTok,
			OutputTokens: outTok,
			TotalTokens:  totTok,
		}

		if err := h.svc.SaveMessage(saveCtx, msg); err != nil {
			zap.L().Error("[friend handler] SaveMessage error", zap.Error(err))
			return
		}

		if isInterrupted {
			zap.L().Info("[friend handler] 对话被中断，跳过记忆总结", zap.Uint("friend_id", req.FriendID))
			return
		}

		count, err := h.svc.GetMessageCount(saveCtx, req.FriendID)
		if err == nil && count > 0 && count%2 == 0 {
			if err := h.svc.UpdateMemory(context.Background(), req.FriendID); err != nil {
				zap.L().Error("[friend handler] 记忆总结失败", zap.Error(err))
			}
		}
	}

	// true 和 false 表示接下来还有无消息
	c.Stream(func(w io.Writer) bool {
		msg, err := stream.Recv()

		// 正常结束
		if err == io.EOF {
			fmt.Println("\n[流式输出结束]")
			c.SSEvent("message", "[DONE]")
			go saveAndUpdateMemory(false, finalOutput, inputTokens, outputTokens, totalTokens)
			return false
		}

		// 发生错误或中断
		if err != nil {
			// 判断是否是客户端主动断开 (Context Canceled)
			if errors.Is(err, context.Canceled) || c.Request.Context().Err() != nil {
				zap.L().Info("[friend handler] 用户主动断开连接", zap.Uint("friendID", req.FriendID))
				go saveAndUpdateMemory(true, finalOutput, inputTokens, outputTokens, totalTokens)
				return false
			}

			// 其他错误
			zap.L().Error("[friend handler] Stream error", zap.Error(err))
			c.SSEvent("error", "生成回复中断")
			go saveAndUpdateMemory(true, finalOutput, inputTokens, outputTokens, totalTokens)
			return false
		}

		// 跳过空字符，直接读取下一个流
		if msg.Content == "" {
			return true
		}

		// 正常拼接输出
		finalOutput += msg.Content
		fmt.Print(msg.Content)

		if msg.ResponseMeta != nil && msg.ResponseMeta.Usage != nil {
			inputTokens = msg.ResponseMeta.Usage.PromptTokens
			outputTokens = msg.ResponseMeta.Usage.CompletionTokens
			totalTokens = msg.ResponseMeta.Usage.TotalTokens
		}

		// 一块一块上传
		c.SSEvent("message", msg.Content)
		return true
	})
}

func (h *friendHandler) GetMessageHistory(c *gin.Context) {
	uid, _ := c.Get("user_id")
	userID, ok := uid.(uint)
	if !ok {
		zap.L().Error("[friend handler] userID type error")
		c.JSON(http.StatusOK, gin.H{"result": "系统繁忙，请稍后再试"})
		return
	}

	friendIDStr := c.Query("friend_id")
	cursorStr := c.Query("cursor")

	friendID, err := strconv.ParseUint(friendIDStr, 10, 32)
	if err != nil || friendID == 0 {
		c.JSON(http.StatusOK, gin.H{"result": "参数格式错误"})
		return
	}
	cursor, err := strconv.ParseUint(cursorStr, 10, 32)
	if err != nil {
		cursor = 0
	}

	msgs, err := h.svc.GetMessageHistory(c.Request.Context(), uint(friendID), userID, uint(cursor), constants.DefaultLimit+1)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"result": err.Error()})
		return
	}

	var hasMore bool
	if len(msgs) > constants.DefaultLimit {
		hasMore = true
		msgs = msgs[:constants.DefaultLimit]
	}

	resps := make([]MessageResp, 0, len(msgs))
	for _, m := range msgs {
		resps = append(resps, MessageResp{
			ID:          m.ID,
			UserMessage: m.UserMessage,
			Output:      m.Output,
		})
	}

	var nextCursor uint
	if len(msgs) > 0 {
		nextCursor = msgs[len(msgs)-1].ID
	}

	c.JSON(http.StatusOK, gin.H{
		"result":      "success",
		"messages":    resps,
		"next_cursor": nextCursor,
		"has_more":    hasMore,
	})
}
