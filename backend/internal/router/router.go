package router

import (
	"backend/internal/character"
	"backend/internal/config"
	"backend/internal/friend"
	"backend/internal/infra/llm"
	"backend/internal/infra/logger"
	"backend/internal/middleware"
	"backend/internal/user"
	"backend/pkg/constants"
	"context"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func SetupRouter(mode string, db *gorm.DB, cfg *config.Config, rdb *redis.Client) *gin.Engine {
	if mode == gin.ReleaseMode {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.New()
	// 配置日志和错误恢复
	r.Use(logger.GinLogger(), logger.GinRecovery(true))
	// 跨域中间件
	r.Use(middleware.CorsMiddleware())
	r.Static("/api/data", "./data")

	userRepo := user.NewUserRepository(db)
	userSvc := user.NewUserService(userRepo)
	userHdl := user.NewUserHandler(userSvc, &cfg.JWT, rdb)

	charRepo := character.NewCharRepository(db)
	charSvc := character.NewCharService(charRepo)
	charHdl := character.NewCharHandler(charSvc)

	ctx := context.Background()
	chatModel, err := llm.InitChatModel(ctx, cfg.AI)
	if err != nil {
		zap.L().Panic("InitChatModel error:", zap.Error(err))
	}
	chatGraph, err := friend.NewChatGraph(ctx, chatModel)
	if err != nil {
		zap.L().Panic("NewChatGraph error:", zap.Error(err))
	}
	memoryGraph, err := friend.NewMemoryGraph(ctx, chatModel)
	if err != nil {
		zap.L().Panic("NewMemoryGraph error:", zap.Error(err))
	}
	friendRepo := friend.NewFriendRepository(db)
	friendSvc := friend.NewFriendService(friendRepo, chatGraph, memoryGraph)
	friendHdl := friend.NewFriendHandler(friendSvc)

	public := r.Group("/api")
	{
		// user
		public.POST("/user/account/register", middleware.RateLimitMiddleware(constants.LimitAuth), userHdl.Register)
		public.POST("/user/account/login", middleware.RateLimitMiddleware(constants.LimitAuth), userHdl.Login)
		public.POST("/user/account/refresh_token", userHdl.RefreshToken)

		// character
		public.GET("/create/character/get_list", charHdl.GetCharList)

		// homepage
		public.GET("/homepage/index", charHdl.GetFeedOrSearch)
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
		protected.POST("/friend/message/chat", middleware.RateLimitMiddleware(constants.LimitChat), friendHdl.StreamChat)
		protected.GET("/friend/message/get_history", friendHdl.GetMessageHistory)
	}

	return r
}
