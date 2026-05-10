package redis

import (
	"backend/pkg/constants"
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
	goredis "github.com/redis/go-redis/v9"
)

var ErrCacheMiss = errors.New(constants.ErrCacheMiss)

type Cache interface {
	Get(ctx context.Context, key string) ([]byte, error)
	Set(ctx context.Context, key string, value any, ttl time.Duration) error
	Delete(ctx context.Context, keys ...string) error
	Exists(ctx context.Context, key string) (bool, error)
	SetNX(ctx context.Context, key string, value any, ttl time.Duration) (bool, error)
	Expire(ctx context.Context, key string, ttl time.Duration) error
	Eval(ctx context.Context, script string, keys []string, args ...any) (any, error)
	Close() error
}

type noopCache struct{}

func NewNoopCache() Cache {
	return noopCache{}
}

func (noopCache) Get(context.Context, string) ([]byte, error) {
	return nil, ErrCacheMiss
}

func (noopCache) Set(context.Context, string, any, time.Duration) error {
	return nil
}

func (noopCache) Delete(context.Context, ...string) error {
	return nil
}

func (noopCache) Exists(context.Context, string) (bool, error) {
	return false, nil
}

func (noopCache) SetNX(context.Context, string, any, time.Duration) (bool, error) {
	return false, nil
}

func (noopCache) Expire(context.Context, string, time.Duration) error {
	return nil
}

func (noopCache) Eval(context.Context, string, []string, ...any) (any, error) {
	return nil, nil
}

func (noopCache) Close() error {
	return nil
}

type redisCache struct {
	client *goredis.Client
}

func NewCache(client *goredis.Client) Cache {
	if client == nil {
		return NewNoopCache()
	}
	return &redisCache{client: client}
}

func (c *redisCache) Get(ctx context.Context, key string) ([]byte, error) {
	data, err := c.client.Get(ctx, key).Bytes()
	if errors.Is(err, goredis.Nil) {
		return nil, ErrCacheMiss
	}
	return data, err
}

func (c *redisCache) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	return c.client.Set(ctx, key, value, ttl).Err()
}

func (c *redisCache) Delete(ctx context.Context, keys ...string) error {
	if len(keys) == 0 {
		return nil
	}
	return c.client.Del(ctx, keys...).Err()
}

func (c *redisCache) Exists(ctx context.Context, key string) (bool, error) {
	count, err := c.client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (c *redisCache) SetNX(ctx context.Context, key string, value any, ttl time.Duration) (bool, error) {
	cmd := c.client.SetArgs(ctx, key, value, redis.SetArgs{
		TTL:  ttl,
		Mode: "NX",
	})
	if cmd.Err() != nil {
		return false, cmd.Err()
	}
	return cmd.Val() == "OK", nil
}

func (c *redisCache) Expire(ctx context.Context, key string, ttl time.Duration) error {
	return c.client.Expire(ctx, key, ttl).Err()
}

func (c *redisCache) Eval(ctx context.Context, script string, keys []string, args ...any) (any, error) {
	return c.client.Eval(ctx, script, keys, args...).Result()
}

func (c *redisCache) Close() error {
	return c.client.Close()
}
