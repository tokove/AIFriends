package middleware

import (
	"backend/internal/config"
	"backend/internal/infra/redis"
	"backend/pkg/utils"
	"net/http"
	"strings"

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

		// 💡 核心新增：Redis 黑名单拦截
		// 只要 Logout 调用过，这里就会拦截掉即便还没过期的 Token
		isBlack, err := redis.RDB.Exists(redis.Ctx, "blacklist:"+tokenString).Result()
		if err == nil && isBlack > 0 {
			c.JSON(http.StatusUnauthorized, gin.H{"result": "登录已失效，请重新登录"})
			c.Abort()
			return
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
