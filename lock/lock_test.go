package lock

import (
	"context"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

func newRDB() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
}

func TestLock(t *testing.T) {
	lc, err := NewClient(newRDB())
	assert.NoError(t, err)

	ctx := context.Background()
	locker1 := lc.NewLocker()

	err = locker1.Unlock(ctx)
	assert.Error(t, err)

	err = locker1.Lock(ctx, "test", 60*time.Second, nil)
	assert.NoError(t, err)
	defer locker1.Unlock(ctx)

	locker2 := lc.NewLocker()
	err = locker2.Lock(ctx, "test", 60*time.Second, nil)
	assert.Error(t, err)
}

func TestTTL(t *testing.T) {
	lc, err := NewClient(newRDB())
	assert.NoError(t, err)

	ctx := context.Background()
	locker := lc.NewLocker()
	err = locker.Lock(ctx, "test", 60*time.Second, nil)
	assert.NoError(t, err)
	defer locker.Unlock(ctx)

	time.Sleep(5 * time.Second)
	d, err := locker.TTL(ctx)
	if assert.NoError(t, err) {
		t.Log(d)
	}
}

func TestSetTTL(t *testing.T) {
	lc, err := NewClient(newRDB())
	assert.NoError(t, err)

	ctx := context.Background()
	locker := lc.NewLocker()
	err = locker.Lock(ctx, "test", 60*time.Second, nil)
	assert.NoError(t, err)
	defer locker.Unlock(ctx)

	time.Sleep(5 * time.Second)
	d, err := locker.TTL(ctx)
	if assert.NoError(t, err) {
		t.Log(d)
	}

	err = locker.SetTTL(ctx, 60*time.Second)
	assert.NoError(t, err)
	d, err = locker.TTL(ctx)
	if assert.NoError(t, err) {
		t.Log(d)
	}
}
