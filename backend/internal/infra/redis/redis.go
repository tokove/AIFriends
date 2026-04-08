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
	Ctx = context.Background()
)

func InitRedis(cfg *config.Config) {
	RDB = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
		PoolSize: cfg.Redis.PoolSize,
	})

	// 设置一个超时 context 检查连接
	ctx, cancel := context.WithTimeout(Ctx, 5*time.Second)
	defer cancel()

	_, err := RDB.Ping(ctx).Result()
	if err != nil {
		zap.L().Fatal("Redis connection failed", zap.Error(err))
	}

	zap.L().Info("Redis connected successfully",
		zap.String("host", cfg.Redis.Host),
		zap.Int("port", cfg.Redis.Port),
	)
}
