package middleware

import (
	"backend/internal/infra/redis"
	"backend/pkg/constants"
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// 如果 key 不存在，初始化为 1 并设置过期时间；否则递增。如果超出限制，返回 0。
const rateLimitLua = `
local current = redis.call("INCR", KEYS[1])
if current == 1 then
    redis.call("EXPIRE", KEYS[1], ARGV[1])
end
if current > tonumber(ARGV[2]) then
    return 0
end
return 1
`

func RateLimitMiddleware(cfg constants.RateLimitConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		if redis.RDB == nil {
			zap.L().Warn("[middleware] Redis 处于离线状态，限流降级放行", zap.String("path", c.FullPath()))
			c.Next()
			return
		}

		// 登录用户用 UID，未登录用户降级用 IP
		var identifier string
		uid, exists := c.Get("user_id")
		if exists {
			identifier = fmt.Sprintf("uid:%v", uid)
		} else {
			identifier = fmt.Sprintf("ip:%s", c.ClientIP())
		}

		// rate_limit:/api/character/create:uid:1001
		rateKey := fmt.Sprintf("rate_limit:%s:%s", c.FullPath(), identifier)

		// 执行 Lua 脚本
		ctx := context.Background()
		result, err := redis.RDB.Eval(ctx, rateLimitLua, []string{rateKey}, cfg.Window, cfg.Max).Result()

		if err != nil {
			zap.L().Error("[middleware] 限流脚本执行失败，降级放行", zap.Error(err))
			c.Next()
			return
		}

		// 判断拦截
		if result.(int64) == 0 {
			zap.L().Warn("[middleware] 触发限流", zap.String("key", rateKey))
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"result": cfg.Msg,
			})
			return
		}

		c.Next()
	}
}
