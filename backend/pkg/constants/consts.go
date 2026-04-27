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
	StaticBaseURL = "/api/media/"

	// 默认路径
	DefaultPath                 = "./configs/config.yaml"
	DefaultUserPhoto            = "user/photos/default.jpg"
	DirUserPhoto                = "user/photos"
	DirCharacterPhoto           = "character/photos"
	DirCharacterBackgroundImage = "character/background_images"
	FrontendDistDir             = "./static/frontend"
	FrontendIndexFile           = "./static/frontend/index.html"

	// 其它
	DefaultLimit        = 20
	DefaultMessageCount = 20
	DefaultRecallLimit  = 200

	// 聊天模块
	SystemPromptTitleReply  = "回复"
	SystemPromptTitleMemory = "记忆"
	ErrSystemBusy           = "系统繁忙，请稍后再试"
	ErrFriendNotFound       = "好友不存在"
	ErrAudioNotFound        = "音频不存在"
	ErrASRFailed            = "语音识别失败"
	ErrCharacterNotFound    = "角色不存在"
	ErrUserNotFound         = "用户不存在"
	ErrTTSFailed            = "语音合成失败"

	MaxChatHistoryCount   = 12
	MaxMemorySummaryCount = 20
	MaxMsgLen             = 500
	MaxContextLength      = 2200
	MaxDBInputLength      = 6000
	ASRChunkSize          = 3200
	MediaCacheMaxAge      = 86400

	AudioTaskGroup          = "audio"
	AudioASRTask            = "asr"
	AudioASRFunction        = "recognition"
	AudioTTSTask            = "tts"
	AudioTTSFunction        = "SpeechSynthesizer"
	AudioStreamingDuplex    = "duplex"
	AudioRunTaskAction      = "run-task"
	AudioContinueTaskAction = "continue-task"
	AudioFinishTaskAction   = "finish-task"
	AudioTaskStartedEvent   = "task-started"
	AudioTaskFinishedEvent  = "task-finished"
	AudioTaskFailedEvent    = "task-failed"
	AudioResultEvent        = "result-generated"

	MarkdownJSONPrefix = "```json\n"
	MarkdownPrefix     = "```\n"
	MarkdownSuffix     = "\n```"
)

type RateLimitConfig struct {
	Window int    // 窗口时间（秒）
	Max    int    // 最大请求数
	Msg    string // 触发后的提示语
}

var (
	LimitChat = RateLimitConfig{
		Window: 60,
		Max:    10,
		Msg:    "好友的 CPU 快冒烟啦，让它稍微喘口气吧～",
	}

	LimitAuth = RateLimitConfig{
		Window: 3600,
		Max:    5,
		Msg:    "操作过于频繁，请稍后再试",
	}

	LimitDiscovery = RateLimitConfig{
		Window: 60,
		Max:    30,
		Msg:    "逛得太快啦！慢下来看看身边有趣的灵魂吧",
	}

	LimitCreateChar = RateLimitConfig{
		Window: 86400,
		Max:    10,
		Msg:    "创建角色额度用光，明天再创建吧",
	}
)
