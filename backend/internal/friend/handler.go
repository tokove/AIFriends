package friend

import (
	"backend/internal/model"
	"backend/pkg/constants"
	"context"
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

	if err := h.svc.DeleteFriend(c.Request.Context(), req.FriendID, userID); err != nil {
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

	stream, inputStr, err := h.svc.StreamChat(c.Request.Context(), req.FriendID, userID, req.Message)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"result": err.Error()})
		return
	}
	defer stream.Close()

	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Flush()

	var finalOutput string
	var inputTokens, outputTokens, totalTokens int

	// 安全的包含中文字符的截断函数
	truncateString := func(str string, maxLen int) string {
		runes := []rune(str)
		if len(runes) > maxLen {
			return string(runes[:maxLen])
		}
		return str
	}

	// 提取出异步存库的逻辑 (将外部变量作为参数传入，避免协程抢占造成的数据竞争)
	saveAndTriggerMemory := func(isInterrupted bool, currentFinalOutput string, inTok, outTok, totTok int) {
		saveCtx := context.Background()

		uMsg := truncateString(req.Message, 500)
		outMsg := truncateString(currentFinalOutput, 500)

		dbMsg := &model.Message{
			FriendID:     req.FriendID,
			UserMessage:  uMsg,
			Input:        inputStr,
			Output:       outMsg,
			InputTokens:  inTok,
			OutputTokens: outTok,
			TotalTokens:  totTok,
		}

		if err := h.svc.SaveMessage(saveCtx, dbMsg); err != nil {
			zap.L().Error("[friend handler] SaveMessage error", zap.Error(err))
		}

		previewMsg := truncateString(currentFinalOutput, 50)
		if len([]rune(currentFinalOutput)) > 50 {
			previewMsg += "..."
		}
		h.svc.UpdateFriendActiveStatus(saveCtx, req.FriendID, previewMsg)

		if isInterrupted {
			zap.L().Info("[friend handler] 对话被中断，跳过本次记忆总结评估", zap.Uint("friend_id", req.FriendID))
			return
		}

		count, err := h.svc.GetMessageCount(saveCtx, req.FriendID)
		if err == nil && count > 0 && count%2 == 0 {
			zap.L().Info("[friend handler] 触发记忆总结", zap.Uint("friend_id", req.FriendID))

			go func() {
				memCtx := context.Background()
				if err := h.svc.UpdateMemory(memCtx, req.FriendID); err != nil {
					zap.L().Error("[friend handler] 记忆总结失败", zap.Error(err))
				} else {
					zap.L().Info("[friend handler] 记忆总结完成并落盘", zap.Uint("friend_id", req.FriendID))
				}
			}()
		}
	}

	c.Stream(func(w io.Writer) bool {
		// 用户中途关掉网页断开连接
		if c.Request.Context().Err() != nil {
			zap.L().Warn("[friend handler] 用户主动断开连接，提前落盘止损", zap.Uint("friendID", req.FriendID))
			// 使用 go 协程去存，立刻释放当前的 Gin 工作线程
			go saveAndTriggerMemory(true, finalOutput, inputTokens, outputTokens, totalTokens)
			return false
		}

		msg, err := stream.Recv()

		if err == io.EOF {
			fmt.Println("\n[流式输出结束]")
			c.SSEvent("message", "[DONE]")
			go saveAndTriggerMemory(false, finalOutput, inputTokens, outputTokens, totalTokens)
			return false
		}

		// 大模型 API 中断，把已经生成的半截话保存下来
		if err != nil {
			zap.L().Error("[friend handler] Stream error", zap.Error(err))
			c.SSEvent("error", "生成回复中断")
			go saveAndTriggerMemory(true, finalOutput, inputTokens, outputTokens, totalTokens)
			return false
		}

		finalOutput += msg.Content
		fmt.Print(msg.Content)

		if msg.ResponseMeta != nil && msg.ResponseMeta.Usage != nil {
			inputTokens = msg.ResponseMeta.Usage.PromptTokens
			outputTokens = msg.ResponseMeta.Usage.CompletionTokens
			totalTokens = msg.ResponseMeta.Usage.TotalTokens
		}

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
	lastMsgIDStr := c.DefaultQuery("last_message_id", "0")

	friendID, _ := strconv.ParseUint(friendIDStr, 10, 32)
	lastMsgID, _ := strconv.ParseUint(lastMsgIDStr, 10, 32)

	if friendID == 0 {
		c.JSON(http.StatusOK, gin.H{"result": "参数格式错误"})
		return
	}

	msgs, err := h.svc.GetMessageHistory(c.Request.Context(), uint(friendID), userID, uint(lastMsgID), constants.DefaultLimit)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"result": err.Error()})
		return
	}

	var resList []MessageResp
	for _, m := range msgs {
		resList = append(resList, MessageResp{
			ID:          m.ID,
			UserMessage: m.UserMessage,
			Output:      m.Output,
		})
	}

	if resList == nil {
		resList = make([]MessageResp, 0)
	}

	c.JSON(http.StatusOK, gin.H{
		"result":   "success",
		"messages": resList,
	})
}
