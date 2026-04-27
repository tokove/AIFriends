package friend

import (
	"backend/internal/model"
	"backend/pkg/constants"
	"context"
	"encoding/base64"
	"errors"
	"io"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type friendHandler struct {
	svc FriendService
}

func NewFriendHandler(svc FriendService) *friendHandler {
	return &friendHandler{svc: svc}
}

func friendErrorStatus(err error) int {
	if err == nil {
		return http.StatusOK
	}

	switch err.Error() {
	case constants.ErrFriendNotFound:
		return http.StatusNotFound
	case constants.ErrAudioNotFound:
		return http.StatusBadRequest
	case constants.ErrASRFailed:
		return http.StatusBadGateway
	case constants.ErrTTSFailed:
		return http.StatusBadGateway
	case constants.ErrSystemBusy:
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}

func (h *friendHandler) buildFriendResp(f *model.Friend) (FriendResp, bool) {
	if f == nil || f.Character == nil || f.Character.Author == nil {
		return FriendResp{}, false
	}
	char := f.Character
	author := char.Author
	return FriendResp{
		ID:        f.ID,
		UpdatedAt: f.UpdatedAt.UnixMilli(),
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
		c.JSON(http.StatusUnauthorized, gin.H{"result": "未登录"})
		return
	}

	var req GetOrCreateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"result": "参数格式错误"})
		return
	}

	friend, err := h.svc.GetOrCreate(c.Request.Context(), req.CharacterID, userID)
	if err != nil {
		c.JSON(friendErrorStatus(err), gin.H{"result": err.Error()})
		return
	}

	friendResp, ok := h.buildFriendResp(friend)
	if !ok {
		zap.L().Debug("[friend handler] buildFriendResp error")
		c.JSON(http.StatusInternalServerError, gin.H{"result": "系统繁忙，请稍后再试"})
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
		c.JSON(http.StatusUnauthorized, gin.H{"result": "未登录"})
		return
	}

	cursorUpdatedAtStr := c.Query("cursor_updated_at")
	cursorIDStr := c.DefaultQuery("cursor_id", "0")

	var cursorUpdatedAt *time.Time
	if cursorUpdatedAtStr != "" {
		cursorUpdatedAtUnix, err := strconv.ParseInt(cursorUpdatedAtStr, 10, 64)
		if err != nil || cursorUpdatedAtUnix <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"result": "参数格式错误"})
			return
		}
		parsedTime := time.UnixMilli(cursorUpdatedAtUnix)
		cursorUpdatedAt = &parsedTime
	}

	cursorID, err := strconv.ParseUint(cursorIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"result": "参数格式错误"})
		return
	}
	if cursorUpdatedAt != nil && cursorID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"result": "参数格式错误"})
		return
	}

	rawFriends, err := h.svc.GetList(c.Request.Context(), userID, cursorUpdatedAt, uint(cursorID))
	if err != nil {
		c.JSON(friendErrorStatus(err), gin.H{"result": err.Error()})
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
		c.JSON(http.StatusUnauthorized, gin.H{"result": "未登录"})
		return
	}

	var req RemoveReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"result": "参数格式错误"})
		return
	}

	if err := h.svc.RemoveFriend(c.Request.Context(), req.FriendID, userID); err != nil {
		c.JSON(friendErrorStatus(err), gin.H{"result": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"result": "success"})
}

func (h *friendHandler) ASR(c *gin.Context) {
	audioFile, err := c.FormFile("audio")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"result": constants.ErrAudioNotFound})
		return
	}

	file, err := audioFile.Open()
	if err != nil {
		zap.L().Error("[friend handler] open audio file failed", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"result": constants.ErrAudioNotFound})
		return
	}
	defer file.Close()

	pcmData, err := io.ReadAll(file)
	if err != nil {
		zap.L().Error("[friend handler] read audio file failed", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"result": constants.ErrAudioNotFound})
		return
	}

	text, err := h.svc.ASR(c.Request.Context(), pcmData)
	if err != nil {
		c.JSON(friendErrorStatus(err), gin.H{"result": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"result": "success",
		"text":   text,
	})
}

func (h *friendHandler) StreamChat(c *gin.Context) {
	uid, _ := c.Get("user_id")
	userID, ok := uid.(uint)
	if !ok {
		zap.L().Error("[friend handler] userID type error in StreamChat")
		c.JSON(http.StatusUnauthorized, gin.H{"result": "未登录"})
		return
	}

	var req ChatReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"result": "参数格式错误"})
		return
	}

	// 获取流和输入
	stream, inputStr, err := h.svc.StreamChat(c.Request.Context(), req.FriendID, userID, req.Message)
	if err != nil {
		c.JSON(friendErrorStatus(err), gin.H{"result": err.Error()})
		return
	}
	defer stream.Close() // 关流

	c.Writer.Header().Set("Content-Type", "text/event-stream") // 设置响应头为 SSE
	c.Writer.Header().Set("Cache-Control", "no-cache")         // 禁止缓存 -> nginx
	c.Writer.Header().Set("X-Accel-Buffering", "no")
	c.Writer.Header().Set("Connection", "keep-alive") // 保持长连接获取信息
	c.Writer.Flush()

	type streamEvent struct {
		Content      string
		AudioBase64  string
		Err          error
		Interrupted  bool
		InputTokens  int
		OutputTokens int
		TotalTokens  int
	}

	var finalOutput string
	var inputTokens, outputTokens, totalTokens int
	ttsInputCh := make(chan string, 16)
	eventCh := make(chan streamEvent, 32)

	audioCh, audioErrCh, ttsErr := h.svc.StreamTTS(c.Request.Context(), req.FriendID, userID, ttsInputCh)
	ttsEnabled := ttsErr == nil
	if ttsErr != nil {
		zap.L().Error("[friend handler] init stream tts failed", zap.Error(ttsErr))
		close(ttsInputCh)
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer func() {
			if ttsEnabled {
				close(ttsInputCh)
			}
		}()

		var curInputTokens, curOutputTokens, curTotalTokens int

		for {
			msg, err := stream.Recv()
			if err == io.EOF {
				return
			}
			if err != nil {
				eventCh <- streamEvent{
					Err:         err,
					Interrupted: errors.Is(err, context.Canceled) || c.Request.Context().Err() != nil,
				}
				return
			}
			if msg.Content == "" {
				continue
			}

			if msg.ResponseMeta != nil && msg.ResponseMeta.Usage != nil {
				curInputTokens = msg.ResponseMeta.Usage.PromptTokens
				curOutputTokens = msg.ResponseMeta.Usage.CompletionTokens
				curTotalTokens = msg.ResponseMeta.Usage.TotalTokens
			}

			if ttsEnabled {
				select {
				case <-c.Request.Context().Done():
					return
				case ttsInputCh <- msg.Content:
				}
			}

			eventCh <- streamEvent{
				Content:      msg.Content,
				InputTokens:  curInputTokens,
				OutputTokens: curOutputTokens,
				TotalTokens:  curTotalTokens,
			}
		}
	}()

	if ttsEnabled {
		wg.Add(1)
		go func() {
			defer wg.Done()
			currentAudioErrCh := audioErrCh
			for {
				select {
				case <-c.Request.Context().Done():
					return
				case err, ok := <-currentAudioErrCh:
					if !ok {
						currentAudioErrCh = nil
						continue
					}
					if err != nil && !errors.Is(err, context.Canceled) {
						zap.L().Error("[friend handler] stream tts failed", zap.Error(err))
					}
					return
				case audioChunk, ok := <-audioCh:
					if !ok {
						return
					}
					eventCh <- streamEvent{
						AudioBase64: base64.StdEncoding.EncodeToString(audioChunk),
					}
				}
			}
		}()
	}

	go func() {
		wg.Wait()
		close(eventCh)
	}()

	// 判断字符串转为 rune 的长度，根据 maxLen 截取
	truncateString := func(str string, maxLen int) string {
		runes := []rune(str)
		if len(runes) > maxLen {
			return string(runes[:maxLen])
		}
		return str
	}

	// 保存消息并检查是否更新记忆（当前按每 2 条触发一次）
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

	c.Stream(func(w io.Writer) bool {
		event, ok := <-eventCh
		if !ok {
			c.SSEvent("message", "[DONE]")
			go saveAndUpdateMemory(false, finalOutput, inputTokens, outputTokens, totalTokens)
			return false
		}

		if event.Err != nil {
			if event.Interrupted {
				zap.L().Info("[friend handler] 用户主动断开连接", zap.Uint("friendID", req.FriendID))
				go saveAndUpdateMemory(true, finalOutput, inputTokens, outputTokens, totalTokens)
				return false
			}

			zap.L().Error("[friend handler] Stream error", zap.Error(event.Err))
			c.SSEvent("error", "生成回复中断")
			go saveAndUpdateMemory(true, finalOutput, inputTokens, outputTokens, totalTokens)
			return false
		}

		if event.Content == "" && event.AudioBase64 == "" {
			return true
		}

		if event.Content != "" {
			finalOutput += event.Content
			inputTokens = event.InputTokens
			outputTokens = event.OutputTokens
			totalTokens = event.TotalTokens
			c.SSEvent("message", gin.H{"content": event.Content})
		}

		if event.AudioBase64 != "" {
			c.SSEvent("message", gin.H{"audio": event.AudioBase64})
		}
		return true
	})
}

func (h *friendHandler) GetMessageHistory(c *gin.Context) {
	uid, _ := c.Get("user_id")
	userID, ok := uid.(uint)
	if !ok {
		zap.L().Error("[friend handler] userID type error")
		c.JSON(http.StatusUnauthorized, gin.H{"result": "未登录"})
		return
	}

	friendIDStr := c.Query("friend_id")
	cursorStr := c.Query("cursor")

	friendID, err := strconv.ParseUint(friendIDStr, 10, 32)
	if err != nil || friendID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"result": "参数格式错误"})
		return
	}
	cursor, err := strconv.ParseUint(cursorStr, 10, 32)
	if err != nil {
		cursor = 0
	}

	msgs, err := h.svc.GetMessageHistory(c.Request.Context(), uint(friendID), userID, uint(cursor), constants.DefaultLimit+1)
	if err != nil {
		c.JSON(friendErrorStatus(err), gin.H{"result": err.Error()})
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
