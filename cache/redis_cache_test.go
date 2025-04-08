package cache

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const (
	key = "test"
	val = "test"
)

func newRedisCache() *RedisCache {
	return NewRedisCache(Config{
		Addr: "localhost:6379",
	})
}

func TestSet(t *testing.T) {
	ctx := context.Background()
	rc := newRedisCache()

	err := rc.Set(ctx, key, val, 60*time.Second)
	assert.NoError(t, err)
}

func TestGet(t *testing.T) {
	ctx := context.Background()
	rc := newRedisCache()

	err := rc.Set(ctx, key, val, 60*time.Second)
	assert.NoError(t, err)

	v, err := rc.Get(ctx, "test")
	if assert.NoError(t, err) {
		assert.Equal(t, val, v)
	}
}
