package middleware

import (
	"backend/internal/config"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func CorsMiddleware(corsCfg config.CorsConfig) gin.HandlerFunc {
	allowedOrigins := make(map[string]struct{}, len(corsCfg.AllowOrigins))
	allowAnyOrigin := false
	for _, origin := range corsCfg.AllowOrigins {
		trimmed := strings.TrimSpace(origin)
		if trimmed == "" {
			continue
		}
		if trimmed == "*" {
			allowAnyOrigin = true
			continue
		}
		allowedOrigins[trimmed] = struct{}{}
	}

	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")

		if origin != "" {
			if allowAnyOrigin {
				if corsCfg.AllowCredentials {
					c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
				} else {
					c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
				}
			} else {
				if _, ok := allowedOrigins[origin]; ok {
					c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
				}
			}
		}

		if corsCfg.AllowCredentials {
			c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		}
		c.Writer.Header().Set("Vary", "Origin")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Platform")
		c.Writer.Header().Set("Access-Control-Expose-Headers", "Content-Length, Authorization")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
