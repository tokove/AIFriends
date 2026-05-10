package constants

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
