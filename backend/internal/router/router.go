package router

import (
	"backend/internal/character"
	"backend/internal/config"
	"backend/internal/infra/logger"
	"backend/internal/middleware"
	"backend/internal/user"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SetupRouter(mode string, db *gorm.DB, cfg *config.Config) *gin.Engine {
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
	userHdl := user.NewUserHandler(userSvc, &cfg.JWT)

	charRepo := character.NewCharRepository(db)
	charSvc := character.NewCharService(charRepo)
	charHdl := character.NewCharHandler(charSvc)

	public := r.Group("/api")
	{
		// user
		public.POST("/user/account/register", userHdl.Register)
		public.POST("/user/account/login", userHdl.Login)
		public.POST("/user/account/refresh_token", userHdl.RefreshToken)

	}

	protected := r.Group("/api")
	protected.Use(middleware.AuthMiddleware(cfg))
	{
		// user
		protected.POST("/user/account/logout", userHdl.Logout)
		protected.GET("/user/account/get_user_info", userHdl.GetUserInfo)
		protected.POST("/user/profile/update/", userHdl.UpdateProfile)

		// character
		protected.POST("/create/character/create", charHdl.CreateChar)
		protected.POST("/create/character/update", charHdl.UpdateChar)
		protected.GET("/create/character/get_single", charHdl.GetCharSingle)
		protected.GET("/create/character/get_list", charHdl.GetCharList)
		protected.POST("/create/character/remove", charHdl.DeleteChar)
	}

	return r
}
