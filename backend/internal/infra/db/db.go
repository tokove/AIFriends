package db

import (
	"backend/internal/config"
	"backend/internal/model"
	"fmt"
	"time"

	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func InitDB(cfg *config.Config) {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%d sslmode=%s TimeZone=%s",
		cfg.DB.Host,
		cfg.DB.User,
		cfg.DB.Password,
		cfg.DB.Name,
		cfg.DB.Port,
		cfg.DB.SSLMode,
		cfg.DB.TimeZone,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		PrepareStmt: true,
		Logger:      logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		zap.L().Fatal("failed to connect database", zap.Error(err))
	}

	sqlDB, err := db.DB()
	if err != nil {
		zap.L().Fatal("failed to get sql.DB from gorm", zap.Error(err))
	}

	sqlDB.SetMaxIdleConns(100)
	sqlDB.SetMaxOpenConns(50)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)
	sqlDB.SetConnMaxIdleTime(10 * time.Minute)

	DB = db
	zap.L().Info("Database connected:", zap.String("db", cfg.DB.Name))
}

func AutoMigrate() {
	if err := DB.AutoMigrate(
		&model.User{},
		&model.Voice{},
		&model.Character{},
		&model.Friend{},
		&model.Message{},
		&model.SystemPrompt{},
	); err != nil {
		zap.L().Fatal("failed to automigrate", zap.Error(err))
	}

	seedVoices()
}

func seedVoices() {
	voices := []model.Voice{
		// 男声
		{Name: "干净清爽男", VoiceID: "longanshuo"},
		{Name: "睿智轻熟男", VoiceID: "longanzhi"},
		{Name: "磁性低音男", VoiceID: "longxiaocheng_v2"},
		// 女声
		{Name: "温婉邻家女", VoiceID: "longxing_v2"},
		{Name: "甜美娇气女", VoiceID: "longfeifei_v2"},
		{Name: "温暖春风女", VoiceID: "longyan_v2"},
	}

	for _, voice := range voices {
		if err := DB.FirstOrCreate(&voice, model.Voice{VoiceID: voice.VoiceID}).Error; err != nil {
			zap.L().Fatal("failed to seed voices", zap.Error(err))
		}
	}
}
