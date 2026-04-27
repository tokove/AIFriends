package router

import (
	"backend/internal/character"
	"backend/internal/config"
	"backend/internal/friend"
	"backend/internal/friend/agent/graph"
	"backend/internal/friend/agent/tool"
	"backend/internal/infra/audio"
	"backend/internal/infra/db"
	"backend/internal/infra/llm"
	"backend/internal/infra/logger"
	"backend/internal/middleware"
	"backend/internal/user"
	"backend/pkg/constants"
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func SetupRouter(mode string, basedb *gorm.DB, cfg *config.Config, rdb *redis.Client) *gin.Engine {
	if mode == gin.ReleaseMode {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.New()
	// 配置日志和错误恢复
	r.Use(logger.GinLogger(), logger.GinRecovery(true))
	// 跨域中间件
	r.Use(middleware.CorsMiddleware(cfg.Cors))
	r.Use(func(c *gin.Context) {
		if c.Request.Method == http.MethodGet && (strings.HasPrefix(c.Request.URL.Path, "/api/media/") || strings.HasPrefix(c.Request.URL.Path, "/api/data/")) {
			c.Header("Cache-Control", fmt.Sprintf("public, max-age=%d", constants.MediaCacheMaxAge))
		}
		c.Next()
	})
	r.Static("/api/media", "./media")
	r.GET("/", func(c *gin.Context) {
		c.File(constants.FrontendIndexFile)
	})

	userRepo := user.NewUserRepository(basedb)
	userSvc := user.NewUserService(userRepo)
	userHdl := user.NewUserHandler(userSvc, &cfg.JWT, rdb)

	charRepo := character.NewCharRepository(basedb)
	charSvc := character.NewCharService(charRepo)
	charHdl := character.NewCharHandler(charSvc)

	ctx := context.Background()
	chatModel, err := llm.InitChatModel(ctx, cfg.Agent)
	if err != nil {
		zap.L().Panic("InitChatModel error:", zap.Error(err))
	}
	embedModel, err := llm.NewDefaultEmbedder(cfg.Agent)
	if err != nil {
		zap.L().Panic("NewDefaultEmbedder error:", zap.Error(err))
	}
	vectordb := db.NewVectorDB(basedb, embedModel)
	tools := tool.InitTools(vectordb)
	chatGraph, err := graph.NewChatGraph(ctx, chatModel, tools)
	if err != nil {
		zap.L().Panic("NewChatGraph error:", zap.Error(err))
	}
	memoryGraph, err := graph.NewMemoryGraph(ctx, chatModel)
	if err != nil {
		zap.L().Panic("NewMemoryGraph error:", zap.Error(err))
	}
	audioSvc := audio.NewService(&cfg.Audio)
	friendRepo := friend.NewFriendRepository(basedb)
	friendSvc := friend.NewFriendService(friendRepo, audioSvc, chatGraph, memoryGraph)
	friendHdl := friend.NewFriendHandler(friendSvc)

	public := r.Group("/api")
	{
		// user
		public.POST("/user/account/register", middleware.RateLimitMiddleware(constants.LimitAuth), userHdl.Register)
		public.POST("/user/account/login", userHdl.Login)
		public.POST("/user/account/refresh_token", userHdl.RefreshToken)

		// character
		public.GET("/create/character/get_list", charHdl.GetCharList)

		// homepage
		public.GET("/homepage/index", charHdl.HomeOrSearch)
	}

	protected := r.Group("/api")
	protected.Use(middleware.AuthMiddleware(cfg))
	{
		// user
		protected.POST("/user/account/logout", userHdl.Logout)
		protected.GET("/user/account/get_user_info", userHdl.GetUserInfo)
		protected.POST("/user/profile/update/", userHdl.UpdateProfile)

		// character
		protected.POST("/create/character/create", middleware.RateLimitMiddleware(constants.LimitCreateChar), charHdl.CreateChar)
		protected.POST("/create/character/update", charHdl.UpdateChar)
		protected.GET("/create/character/get_single", charHdl.GetCharSingle)
		protected.POST("/create/character/remove", charHdl.DeleteChar)

		// friend
		protected.POST("/friend/get_or_create", friendHdl.GetOrCreate)
		protected.GET("/friend/get_list", friendHdl.GetFriendList)
		protected.POST("/friend/remove", friendHdl.RemoveFriend)
		protected.POST("/friend/message/asr", middleware.RateLimitMiddleware(constants.LimitChat), friendHdl.ASR)
		protected.POST("/friend/message/chat", middleware.RateLimitMiddleware(constants.LimitChat), friendHdl.StreamChat)
		protected.GET("/friend/message/get_history", friendHdl.GetMessageHistory)
	}

	r.NoRoute(func(c *gin.Context) {
		requestPath := c.Request.URL.Path
		if requestPath == "/api" || strings.HasPrefix(requestPath, "/api/") {
			c.JSON(http.StatusNotFound, gin.H{"result": "接口不存在"})
			return
		}

		cleanPath := filepath.Clean(strings.TrimPrefix(requestPath, "/"))
		if cleanPath == "." {
			c.File(constants.FrontendIndexFile)
			return
		}

		staticFilePath := filepath.Join(constants.FrontendDistDir, cleanPath)
		if fileInfo, err := os.Stat(staticFilePath); err == nil && !fileInfo.IsDir() {
			c.File(staticFilePath)
			return
		}

		c.File(constants.FrontendIndexFile)
	})

	return r
}
