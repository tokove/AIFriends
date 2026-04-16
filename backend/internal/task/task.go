package task

import (
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func StartChatCountResetTask(db *gorm.DB) {
	c := cron.New(cron.WithSeconds())

	// 秒 分 时 日 月 周
	_, err := c.AddFunc("0 0 0 * * *", func() {
		zap.L().Info("开始执行每日 0 点重置聊天计数任务...")
		if err := db.Exec("UPDATE characters SET recent_chat_count = 0 WHERE recent_chat_count != 0").Error; err != nil {
			zap.L().Error("重置聊天计数失败", zap.Error(err))
		}
		zap.L().Info("重置聊天计数成功")
	})
	if err != nil {
		zap.L().Fatal("定时任务解析失败", zap.Error(err))
	}

	c.Start()
	zap.L().Info("后台定时任务调度器已启动")
}
