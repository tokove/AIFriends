package character

import (
	"backend/pkg/constants"
	"backend/pkg/utils"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type charHandler struct {
	svc CharService
}

func NewCharHandler(svc CharService) *charHandler {
	return &charHandler{svc: svc}
}

func charErrorStatus(err error) int {
	if err == nil {
		return http.StatusOK
	}

	msg := err.Error()
	switch {
	case msg == "角色不存在":
		return http.StatusNotFound
	case msg == "用户不存在":
		return http.StatusNotFound
	case msg == "介绍不能为空":
		return http.StatusBadRequest
	case strings.HasPrefix(msg, "名字长度需在"):
		return http.StatusBadRequest
	case strings.HasPrefix(msg, "介绍太长了"):
		return http.StatusBadRequest
	case msg == "头像上传失败":
		return http.StatusInternalServerError
	case msg == "背景图片上传失败":
		return http.StatusInternalServerError
	case msg == "新头像上传失败":
		return http.StatusInternalServerError
	case msg == "新背景图片上传失败":
		return http.StatusInternalServerError
	case msg == "更新失败，请稍后再试":
		return http.StatusInternalServerError
	case msg == "系统繁忙，请稍后再试":
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}

func (h *charHandler) CreateChar(c *gin.Context) {
	uid, _ := c.Get("user_id")
	userID, ok := uid.(uint)
	if !ok {
		zap.L().Error("[char handler] userID type error")
		c.JSON(http.StatusUnauthorized, gin.H{"result": "未登录"})
		return
	}

	name := c.PostForm("name")
	voiceID := c.PostForm("voice_id")
	profile := c.PostForm("profile")

	photo, err := c.FormFile("photo")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"result": "请上传角色头像"})
		return
	}
	if err := utils.CheckImage(photo); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"result": err.Error()})
		return
	}

	bg, err := c.FormFile("background_image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"result": "请上传角色背景图片"})
		return
	}
	if err := utils.CheckImage(bg); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"result": err.Error()})
		return
	}

	if err := h.svc.CreateChar(c.Request.Context(), userID, name, voiceID, profile, photo, bg); err != nil {
		c.JSON(charErrorStatus(err), gin.H{"result": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"result": "success"})
}

func (h *charHandler) UpdateChar(c *gin.Context) {
	uid, _ := c.Get("user_id")
	userID, ok := uid.(uint)
	if !ok {
		zap.L().Error("[char handler] userID type error")
		c.JSON(http.StatusUnauthorized, gin.H{"result": "未登录"})
		return
	}

	cid := c.PostForm("character_id")
	charID, err := strconv.ParseUint(cid, 10, 64)
	if err != nil {
		zap.L().Error("[char handler] charID type error")
		c.JSON(http.StatusBadRequest, gin.H{"result": "参数格式错误"})
		return
	}

	name := c.PostForm("name")
	voiceID := c.PostForm("voice_id")
	profile := c.PostForm("profile")

	photo, err := c.FormFile("photo")
	if err != nil {
		if !errors.Is(err, http.ErrMissingFile) {
			c.JSON(http.StatusBadRequest, gin.H{"result": "图片数据异常"})
			return
		}
		photo = nil
	}
	if photo != nil {
		if err := utils.CheckImage(photo); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"result": err.Error()})
			return
		}
	}

	bg, err := c.FormFile("background_image")
	if err != nil {
		if !errors.Is(err, http.ErrMissingFile) {
			c.JSON(http.StatusBadRequest, gin.H{"result": "图片数据异常"})
			return
		}
		bg = nil
	}
	if bg != nil {
		if err := utils.CheckImage(bg); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"result": err.Error()})
			return
		}
	}

	if err := h.svc.UpdateChar(c.Request.Context(), userID, uint(charID), name, voiceID, profile, photo, bg); err != nil {
		c.JSON(charErrorStatus(err), gin.H{"result": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"result": "success"})
}

func (h *charHandler) GetCharSingle(c *gin.Context) {
	charIDStr := c.Query("character_id")
	charID, err := strconv.ParseUint(charIDStr, 10, 64)
	if err != nil {
		zap.L().Error("[char handler] ParseUint error", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"result": "参数格式错误"})
		return
	}
	if charID == 0 {
		zap.L().Error("[char handler] charID value is zero")
		c.JSON(http.StatusBadRequest, gin.H{"result": "参数格式错误"})
		return
	}

	char, err := h.svc.GetCharSingle(c.Request.Context(), uint(charID))
	if err != nil {
		c.JSON(charErrorStatus(err), gin.H{"result": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"result":    "success",
		"character": char,
	})
}

func (h *charHandler) GetCharList(c *gin.Context) {
	userIDStr := c.Query("user_id")
	cursorUpdatedAtStr := c.Query("cursor_updated_at")
	cursorIDStr := c.DefaultQuery("cursor_id", "0")

	userID, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		zap.L().Error("[char handler] ParseUint error", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"result": "参数格式错误"})
		return
	}
	if userID == 0 {
		zap.L().Error("[char handler] userID value is zero")
		c.JSON(http.StatusBadRequest, gin.H{"result": "参数格式错误"})
		return
	}

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

	cursorID, err := strconv.ParseUint(cursorIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"result": "参数格式错误"})
		return
	}
	if cursorUpdatedAt != nil && cursorID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"result": "参数格式错误"})
		return
	}

	profileUser, err := h.svc.GetUserProfile(c.Request.Context(), uint(userID))
	if err != nil {
		c.JSON(charErrorStatus(err), gin.H{"result": err.Error()})
		return
	}

	rawChars, err := h.svc.GetUserChars(c.Request.Context(), uint(userID), cursorUpdatedAt, uint(cursorID))
	if err != nil {
		c.JSON(charErrorStatus(err), gin.H{"result": err.Error()})
		return
	}

	user := UserProfileResp{
		UserID:   profileUser.ID,
		Username: profileUser.Username,
		Profile:  profileUser.Profile,
		Photo:    constants.StaticBaseURL + profileUser.Photo,
	}

	fallbackAuthor := &AuthorInfoResp{
		UserID:   profileUser.ID,
		Username: profileUser.Username,
		Photo:    constants.StaticBaseURL + profileUser.Photo,
	}

	chars := make([]CharacterItemResp, 0, len(rawChars))
	for _, char := range rawChars {
		if char == nil {
			continue
		}

		authorInfo := AuthorInfoResp{}
		if char.Author != nil {
			authorInfo = AuthorInfoResp{
				UserID:   char.Author.ID,
				Username: char.Author.Username,
				Photo:    constants.StaticBaseURL + char.Author.Photo,
			}
		} else {
			authorInfo = *fallbackAuthor
		}

		item := CharacterItemResp{
			ID:              char.ID,
			UpdatedAt:       char.UpdatedAt.UnixMilli(),
			Name:            char.Name,
			VoiceID:         char.VoiceID,
			Profile:         char.Profile,
			Photo:           constants.StaticBaseURL + char.Photo,
			BackgroundImage: constants.StaticBaseURL + char.BackgroundImage,
			Author:          authorInfo,
		}
		chars = append(chars, item)
	}

	c.JSON(http.StatusOK, gin.H{
		"result":       "success",
		"user_profile": user,
		"characters":   chars,
	})
}

func (h *charHandler) DeleteChar(c *gin.Context) {
	uid, _ := c.Get("user_id")
	userID, ok := uid.(uint)
	if !ok {
		zap.L().Error("[char handler] userID type error")
		c.JSON(http.StatusUnauthorized, gin.H{"result": "未登录"})
		return
	}

	var req DeleteCharReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"result": "参数格式错误"})
		return
	}

	if err := h.svc.DeleteChar(c.Request.Context(), userID, req.CharID); err != nil {
		c.JSON(charErrorStatus(err), gin.H{"result": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"result": "success"})
}

func (h *charHandler) HomeOrSearch(c *gin.Context) {
	query := c.Query("query")
	cursor := c.Query("cursor")

	var cursorTime int64 = 0
	var cursorID uint = 0
	var err error

	if cursor != "" {
		parts := strings.Split(cursor, "_")
		if len(parts) != 2 {
			c.JSON(http.StatusBadRequest, gin.H{"result": "参数格式错误"})
			return
		}

		cursorTime, err = strconv.ParseInt(parts[0], 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"result": "参数格式错误"})
			return
		}
		id, err := strconv.ParseUint(parts[1], 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"result": "参数格式错误"})
			return
		}
		cursorID = uint(id)
	}

	chars, err := h.svc.HomeOrSearch(c.Request.Context(), query, cursorTime, cursorID, constants.DefaultLimit+1)
	if err != nil {
		zap.L().Error("[char handler] GetFeedOrSearch error", zap.Error(err))
		c.JSON(charErrorStatus(err), gin.H{"result": err.Error()})
		return
	}

	var hasMore bool
	if len(chars) > constants.DefaultLimit {
		hasMore = true
		chars = chars[:constants.DefaultLimit]
	}

	var nextCursor string
	if len(chars) > 0 {
		char := chars[len(chars)-1]
		nextCursor = fmt.Sprintf("%d_%d", char.UpdatedAt.Unix(), char.ID)
	}

	c.JSON(http.StatusOK, gin.H{
		"result":      "success",
		"characters":  chars,
		"next_cursor": nextCursor,
		"has_more":    hasMore,
	})
}
