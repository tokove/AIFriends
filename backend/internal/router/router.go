package router

import (
	"backend/internal/infra/logger"
	"backend/internal/middleware"
	"net/http"

	"github.com/gin-gonic/gin"
)

func SetupRouter(mode string) *gin.Engine {
	if mode == gin.ReleaseMode {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.New()

	// 配置日志和错误恢复
	r.Use(logger.GinLogger(), logger.GinRecovery(true))

	// 跨域中间件
	r.Use(middleware.CorsMiddleware())

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	return r
}
