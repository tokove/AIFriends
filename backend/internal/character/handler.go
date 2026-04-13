package character

import (
	"backend/internal/model"
	"backend/pkg/constants"
	"backend/pkg/utils"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type charHandler struct {
	svc CharService
}

func NewCharHandler(svc CharService) *charHandler {
	return &charHandler{svc: svc}
}

func (h *charHandler) CreateChar(c *gin.Context) {
	uid, _ := c.Get("user_id")
	userID, ok := uid.(uint)
	if !ok {
		zap.L().Error("[char handler] userID type error")
		c.JSON(http.StatusOK, gin.H{"result": "系统繁忙，请稍后再试"})
		return
	}

	name := c.PostForm("name")
	profile := c.PostForm("profile")

	photo, err := c.FormFile("photo")
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"result": "请上传角色头像"})
		return
	}
	if err := utils.CheckImage(photo); err != nil {
		c.JSON(http.StatusOK, gin.H{"result": err.Error()})
		return
	}

	bg, err := c.FormFile("background_image")
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"result": "请上传角色背景图片"})
		return
	}
	if err := utils.CheckImage(bg); err != nil {
		c.JSON(http.StatusOK, gin.H{"result": err.Error()})
		return
	}

	if err := h.svc.CreateChar(c.Request.Context(), userID, name, profile, photo, bg); err != nil {
		c.JSON(http.StatusOK, gin.H{"result": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"result": "success"})
}

func (h *charHandler) UpdateChar(c *gin.Context) {
	uid, _ := c.Get("user_id")
	userID, ok := uid.(uint)
	if !ok {
		zap.L().Error("[char handler] userID type error")
		c.JSON(http.StatusOK, gin.H{"result": "系统繁忙，请稍后再试"})
		return
	}

	cid := c.PostForm("character_id")
	charID, err := strconv.ParseUint(cid, 10, 64)
	if err != nil {
		zap.L().Error("[char handler] charID type error")
		c.JSON(http.StatusOK, gin.H{"result": "系统繁忙，请稍后再试"})
		return
	}

	name := c.PostForm("name")
	profile := c.PostForm("profile")

	photo, err := c.FormFile("photo")
	if err != nil {
		if !errors.Is(err, http.ErrMissingFile) {
			c.JSON(http.StatusOK, gin.H{"result": "图片数据异常"})
			return
		}
		photo = nil
	}
	if photo != nil {
		if err := utils.CheckImage(photo); err != nil {
			c.JSON(http.StatusOK, gin.H{"result": err.Error()})
			return
		}
	}

	bg, err := c.FormFile("background_image")
	if err != nil {
		if !errors.Is(err, http.ErrMissingFile) {
			c.JSON(http.StatusOK, gin.H{"result": "图片数据异常"})
			return
		}
		bg = nil
	}
	if bg != nil {
		if err := utils.CheckImage(bg); err != nil {
			c.JSON(http.StatusOK, gin.H{"result": err.Error()})
			return
		}
	}

	if err := h.svc.UpdateChar(c.Request.Context(), userID, uint(charID), name, profile, photo, bg); err != nil {
		c.JSON(http.StatusOK, gin.H{"result": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"result": "success"})
}

func (h *charHandler) GetCharSingle(c *gin.Context) {
	charIDStr := c.DefaultQuery("character_id", "0")
	charID, err := strconv.ParseUint(charIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"result": "参数格式错误"})
		return
	}

	char, err := h.svc.GetCharSingle(c.Request.Context(), uint(charID))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"result": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"result":    "success",
		"character": char,
	})
}

func (h *charHandler) GetCharList(c *gin.Context) {
	userIDStr := c.DefaultQuery("user_id", "0")
	itemsCountStr := c.DefaultQuery("items_count", "0")

	userID, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		zap.L().Error("[char handler] ParseUint error", zap.Error(err))
		c.JSON(http.StatusOK, gin.H{"result": "参数格式错误"})
		return
	}
	itemsCount, err := strconv.ParseInt(itemsCountStr, 10, 64)
	if err != nil {
		zap.L().Error("[char handler] ParseInt error", zap.Error(err))
		c.JSON(http.StatusOK, gin.H{"result": "参数格式错误"})
		return
	}

	if itemsCount < 0 {
		itemsCount = 0
	}
	rawChars, err := h.svc.GetUserChars(c.Request.Context(), uint(userID), int(itemsCount))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"result": err.Error()})
		return
	}

	var author *model.User
	if len(rawChars) > 0 {
		author = rawChars[0].Author
	}
	user := UserProfileResp{
		UserID:   uint(userID),
		Username: author.Username,
		Profile:  author.Profile,
		Photo:    constants.StaticBaseURL + author.Photo,
	}

	chars := make([]CharacterItemResp, 0, len(rawChars))
	for _, char := range rawChars {
		authorInfo := AuthorInfoResp{
			UserID:   author.ID,
			Username: author.Username,
			Photo:    constants.StaticBaseURL + author.Photo,
		}
		item := CharacterItemResp{
			ID:              char.ID,
			Name:            char.Name,
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
		c.JSON(http.StatusOK, gin.H{"result": "系统繁忙，请稍后再试"})
		return
	}

	var req DeleteCharReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"result": "参数格式错误"})
		return
	}

	if err := h.svc.DeleteChar(c.Request.Context(), userID, req.CharID); err != nil {
		c.JSON(http.StatusOK, gin.H{"result": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"result": "success"})
}
