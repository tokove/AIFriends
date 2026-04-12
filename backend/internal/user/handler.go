package user

import (
	"backend/internal/config"
	"backend/internal/model"
	"backend/pkg/constants"
	"backend/pkg/utils"
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type userHandler struct {
	svc UserService
	jwt *config.JwtConfig
	rdb *redis.Client
}

func NewUserHandler(svc UserService, jwt *config.JwtConfig, rdb *redis.Client) *userHandler {
	return &userHandler{
		svc: svc,
		jwt: jwt,
		rdb: rdb,
	}
}

func (h *userHandler) formatUserResponse(user *model.User) gin.H {
	photo := user.Photo
	if photo == "" {
		photo = constants.DefaultUserPhoto
	}
	return gin.H{
		"user_id":  user.ID,
		"username": user.Username,
		"profile":  user.Profile,
		"photo":    constants.StaticBaseURL + photo, // 统一使用常量前缀
	}
}

func (h *userHandler) Register(c *gin.Context) {
	var req RegisterReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"result": "参数格式错误"})
		return
	}

	user, err := h.svc.Register(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"result": err.Error()})
		return
	}

	access, refresh, err := utils.GenerateToken(user.ID, h.jwt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"result": "生成令牌失败"})
		return
	}

	// 参数：name, value, maxAge, path, domain, secure, httpOnly
	c.SetCookie("refresh_token", refresh, h.jwt.RefreshExp, "/", "", true, true)

	resp := h.formatUserResponse(user)
	resp["result"] = "success"
	resp["access"] = access
	c.JSON(http.StatusOK, resp)
}

func (h *userHandler) Login(c *gin.Context) {
	var req LoginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"result": "参数格式错误"})
		return
	}

	user, err := h.svc.Login(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"result": err.Error()})
		return
	}

	access, refresh, err := utils.GenerateToken(user.ID, h.jwt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"result": "生成令牌失败"})
		return
	}

	// 参数：name, value, maxAge, path, domain, secure, httpOnly
	c.SetCookie("refresh_token", refresh, h.jwt.RefreshExp, "/", "", true, true)

	resp := h.formatUserResponse(user)
	resp["result"] = "success"
	resp["access"] = access
	c.JSON(http.StatusOK, resp)
}

func (h *userHandler) Logout(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if authHeader != "" && h.rdb != nil {
		token := strings.TrimPrefix(authHeader, "Bearer ")
		ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
		defer cancel()

		if err := h.rdb.Set(ctx, "blacklist:"+token, "1", time.Duration(h.jwt.AccessExp)*time.Second).Err(); err != nil {
			zap.L().Warn("failed to blacklist token", zap.Error(err))
		}
	}

	c.SetCookie("refresh_token", "", -1, "/", "", true, true)
	c.JSON(http.StatusOK, gin.H{"result": "success"})
}

func (h *userHandler) RefreshToken(c *gin.Context) {
	refreshToken, err := c.Cookie("refresh_token")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"result": "refresh token 不存在"})
		return
	}

	claims, err := utils.ParseToken(refreshToken, h.jwt.Secret)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"result": err.Error()})
		return
	}

	newAccess, newRefresh, err := utils.GenerateToken(claims.UserID, h.jwt)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"result": err.Error()})
		return
	}

	c.SetCookie("refresh_token", newRefresh, h.jwt.RefreshExp, "/", "", true, true)

	c.JSON(http.StatusOK, gin.H{
		"result": "success",
		"access": newAccess,
	})
}

func (h *userHandler) GetUserInfo(c *gin.Context) {
	uid, _ := c.Get("user_id")
	userID, ok := uid.(uint)
	if !ok {
		zap.L().Error("[user handler] userID类型错误")
		c.JSON(http.StatusOK, gin.H{"result": "系统繁忙，请稍后再试"})
		return
	}

	user, err := h.svc.GetUserInfo(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"result": err.Error()})
		return
	}

	resp := h.formatUserResponse(user)
	resp["result"] = "success"
	c.JSON(http.StatusOK, resp)
}

func (h *userHandler) UpdateProfile(c *gin.Context) {
	uid, _ := c.Get("user_id")
	userID, ok := uid.(uint)
	if !ok {
		zap.L().Error("[user handler] userID类型错误")
		c.JSON(http.StatusOK, gin.H{"result": "系统繁忙，请稍后再试"})
		return
	}

	username := c.PostForm("username")
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

	user, err := h.svc.UpdateProfile(c.Request.Context(), userID, username, profile, photo)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"result": err.Error()})
		return
	}

	resp := h.formatUserResponse(user)
	resp["result"] = "success"
	c.JSON(http.StatusOK, resp)
}
