package cache

import (
	"context"
	"errors"
	"time"

	"github.com/eko/gocache/lib/v4/cache"
	"github.com/eko/gocache/lib/v4/store"
	redisstore "github.com/eko/gocache/store/redis/v4"
	"github.com/redis/go-redis/v9"
)

type RedisCacheConfig struct {
	Addr string
}

type RedisCache struct {
	cache *cache.Cache[string]
}

func NewRedisCache(c RedisCacheConfig) *RedisCache {
	rs := redisstore.NewRedis(redis.NewClient(&redis.Options{
		Addr: c.Addr,
	}))

	return &RedisCache{
		cache: cache.New[string](rs),
	}
}

func (rc *RedisCache) Set(ctx context.Context, key string, val string, exp time.Duration) error {
	return rc.cache.Set(ctx, key, val, store.WithExpiration(exp))
}

func (rc *RedisCache) Get(ctx context.Context, key string) (string, error) {
	val, err := rc.cache.Get(ctx, key)
	if err != nil {
		return "", err
	}
	if val == "" {
		return "", errors.New("empty value")
	}

	return val, nil
}
