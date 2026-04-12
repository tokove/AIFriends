package redis

import (
	"context"
	"fmt"
	"time"

	"backend/internal/config"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

var (
	RDB *redis.Client
)

func InitRedis(cfg *config.Config) {
	RDB = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
		PoolSize: cfg.Redis.PoolSize,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_, err := RDB.Ping(ctx).Result()
	if err != nil {
		zap.L().Error("Redis is not available", zap.Error(err))
		RDB = nil
		return
	}

	zap.L().Info("Redis connected successfully")
}
