package middleware

import (
	"backend/internal/config"
	"backend/internal/infra/redis"
	"backend/pkg/utils"
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. 获取并校验格式
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

		tokenString := parts[1]

		if redis.RDB != nil {
			// 设置超时，防止中间件因为 Redis 响应慢而卡死
			ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
			defer cancel()

			isBlack, err := redis.RDB.Exists(ctx, "blacklist:"+tokenString).Result()
			if err == nil && isBlack > 0 {
				c.JSON(http.StatusUnauthorized, gin.H{"result": "登录已失效"})
				c.Abort()
				return
			}
		}

		// 2. 解析 Access Token
		claims, err := utils.ParseToken(tokenString, cfg.JWT.Secret)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"result": "Token无效或已过期"})
			c.Abort()
			return
		}

		// 3. 存入上下文
		c.Set("user_id", claims.UserID)
		c.Next()
	}
}
