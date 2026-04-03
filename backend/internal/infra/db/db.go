package db

import (
	"backend/internal/config"
	"fmt"

	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func InitDB(cfg *config.Config) {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%d sslmode=%s",
		cfg.DB.Host,
		cfg.DB.User,
		cfg.DB.Password,
		cfg.DB.Name,
		cfg.DB.Port,
		cfg.DB.SSLMode,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		zap.L().Fatal("failed to connect database", zap.Error(err))
	}

	DB = db
	zap.L().Info("Database connected:", zap.String("db", cfg.DB.Name))
}

// func AutoMigrate() {
// 	if err := DB.AutoMigrate(); err != nil {
// 		log.Fatalf("db automigrate failed, err: %v", err)
// 	}
// }
