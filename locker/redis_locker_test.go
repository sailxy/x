package locker

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLock(t *testing.T) {
	ctx := context.Background()
	locker := NewLocker(Config{
		Addr: "localhost:6379",
	})

	lock, err := locker.Lock(ctx, "test", 60*time.Second, nil)
	assert.NoError(t, err)
	defer func() { _ = lock.Unlock(ctx) }()

	_, err = locker.Lock(ctx, "test", 60*time.Second, nil)
	assert.Error(t, err)
}

func TestTTL(t *testing.T) {
	ctx := context.Background()
	locker := NewLocker(Config{
		Addr: "localhost:6379",
	})

	lock, err := locker.Lock(ctx, "test", 60*time.Second, nil)
	assert.NoError(t, err)
	defer func() { _ = lock.Unlock(ctx) }()

	time.Sleep(5 * time.Second)
	d, err := lock.TTL(ctx)
	if assert.NoError(t, err) {
		t.Log(d)
	}
}

func TestSetTTL(t *testing.T) {
	ctx := context.Background()
	locker := NewLocker(Config{
		Addr: "localhost:6379",
	})

	lock, err := locker.Lock(ctx, "test", 60*time.Second, nil)
	assert.NoError(t, err)
	defer func() { _ = lock.Unlock(ctx) }()

	time.Sleep(5 * time.Second)
	d, err := lock.TTL(ctx)
	if assert.NoError(t, err) {
		t.Log(d)
	}

	err = lock.SetTTL(ctx, 60*time.Second)
	assert.NoError(t, err)
	d, err = lock.TTL(ctx)
	if assert.NoError(t, err) {
		t.Log(d)
	}
}
