package constants

const (
	// 用户校验限制
	MinUsernameLen    = 2
	MaxUsernameLen    = 32
	MinPasswordLen    = 8
	MaxPasswordLen    = 72
	MaxUserProfileLen = 500
	MinCharNameLen    = 2
	MaxCharNameLen    = 32
	MaxCharProfileLen = 100000

	// 文件相关
	MaxFileSize   = 2 * 1024 * 1024 // 2MB
	StaticBaseURL = "/api/data/"

	// 默认路径
	DefaultUserPhoto            = "user/photos/default.jpg"
	DirUserPhoto                = "user/photos"
	DirCharacterPhoto           = "character/photos"
	DirCharacterBackgroundImage = "character/background_images"

	// 其它
	DefaultLimit        = 20
	DefaultMessageCount = 20

	// 聊天模块
	SystemPromptTitleReply  = "回复"
	SystemPromptTitleMemory = "记忆"

	MaxChatHistoryCount   = 20
	MaxMemorySummaryCount = 20
	MaxContextLength      = 4000  // 上下文防爆字符阈值
	MaxDBInputLength      = 10000 // 数据库存储前强制截断阈值

	MarkdownJSONPrefix = "```json\n"
	MarkdownPrefix     = "```\n"
	MarkdownSuffix     = "\n```"
)
