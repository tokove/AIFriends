package middleware

import (
	"backend/internal/config"
	"backend/pkg/utils"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. 从 Authorization Header 获取 Token: "Bearer <token>"
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"result": "未登录"})
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			c.JSON(http.StatusUnauthorized, gin.H{"result": "Token格式错误"})
			c.Abort()
			return
		}

		// 2. 解析解析 Access Token
		claims, err := utils.ParseToken(parts[1], cfg.JWT.Secret)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"result": "Token无效或已过期"})
			c.Abort()
			return
		}

		// 3. 将解析出的 UserID 存入 Context，方便后续 Handler 使用
		c.Set("user_id", claims.UserID)
		c.Next()
	}
}
