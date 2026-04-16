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

	// 传入的 c.Request.Context() 会在客户端断开时被取消
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

	truncateString := func(str string, maxLen int) string {
		runes := []rune(str)
		if len(runes) > maxLen {
			return string(runes[:maxLen])
		}
		return str
	}

	// 提取出异步存库的逻辑 (去掉了内部多余的 go func)
	saveAndTriggerMemory := func(isInterrupted bool, currentFinalOutput string, inTok, outTok, totTok int) {
		saveCtx := context.Background()

		dbMsg := &model.Message{
			FriendID:     req.FriendID,
			UserMessage:  truncateString(req.Message, 500),
			Input:        inputStr,
			Output:       truncateString(currentFinalOutput, 500),
			InputTokens:  inTok,
			OutputTokens: outTok,
			TotalTokens:  totTok,
		}

		if err := h.svc.SaveMessage(saveCtx, dbMsg); err != nil {
			zap.L().Error("[friend handler] SaveMessage error", zap.Error(err))
			return
		}

		if isInterrupted {
			zap.L().Info("[friend handler] 对话被中断，跳过记忆总结", zap.Uint("friend_id", req.FriendID))
			return
		}

		count, err := h.svc.GetMessageCount(saveCtx, req.FriendID)
		if err == nil && count > 0 && count%2 == 0 {
			// 这里不需要 go func() 了，因为外层已经是 go 调用
			if err := h.svc.UpdateMemory(context.Background(), req.FriendID); err != nil {
				zap.L().Error("[friend handler] 记忆总结失败", zap.Error(err))
			}
		}
	}

	c.Stream(func(w io.Writer) bool {
		msg, err := stream.Recv()

		// 1. 正常结束
		if err == io.EOF {
			fmt.Println("\n[流式输出结束]")
			c.SSEvent("message", "[DONE]")
			go saveAndTriggerMemory(false, finalOutput, inputTokens, outputTokens, totalTokens)
			return false
		}

		// 2. 发生错误或中断
		if err != nil {
			// 判断是否是客户端主动断开 (Context Canceled)
			if errors.Is(err, context.Canceled) || c.Request.Context().Err() != nil {
				zap.L().Info("[friend handler] 用户主动断开连接，提前落盘止损", zap.Uint("friendID", req.FriendID))
				go saveAndTriggerMemory(true, finalOutput, inputTokens, outputTokens, totalTokens)
				return false
			}

			// 真正的未知流式错误
			zap.L().Error("[friend handler] Stream error", zap.Error(err))
			c.SSEvent("error", "生成回复中断") // 仅在真正的错误时发送
			go saveAndTriggerMemory(true, finalOutput, inputTokens, outputTokens, totalTokens)
			return false
		}

		// 3. 正常拼接输出
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
