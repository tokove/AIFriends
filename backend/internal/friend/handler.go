package friend

import (
	"backend/internal/model"
	"backend/pkg/constants"
	"fmt"
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
