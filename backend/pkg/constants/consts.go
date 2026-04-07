package constants

const (
	// 用户校验限制
	MinUsernameLen = 2
	MaxUsernameLen = 32
	MinPasswordLen = 8
	MaxPasswordLen = 72
	MaxProfileLen  = 500

	// 文件相关
	MaxFileSize   = 2 * 1024 * 1024 // 2MB
	StaticBaseURL = "/api/data/"

	// 默认路径
	DefaultUserPhoto            = "user/photos/default.jpg"
	DirUserPhoto                = "user/photos"
	DirCharacterPhoto           = "character/photos"
	DirCharacterBackgroundImage = "character/background_images"
)
