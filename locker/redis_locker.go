package locker

import (
	"context"
	"fmt"
	"time"

	"github.com/bsm/redislock"
	"github.com/redis/go-redis/v9"
)

type Config struct {
	Addr string
}

type LockOptions struct{}

type Locker struct {
	client *redislock.Client
}

func NewLocker(c Config) *Locker {
	client := redis.NewClient(&redis.Options{
		Addr: c.Addr,
	})
	return &Locker{
		client: redislock.New(client),
	}
}

// Try to obtain a lock.
func (l *Locker) Lock(ctx context.Context, key string, ttl time.Duration, opts *LockOptions) (*lock, error) {
	lc, err := l.client.Obtain(ctx, key, ttl, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to obtain lock: %w", err)
	}
	return &lock{l: lc}, nil
}

type lock struct {
	l *redislock.Lock
}

// Check the remaining TTL of the lock.
func (l *lock) TTL(ctx context.Context) (time.Duration, error) {
	return l.l.TTL(ctx)
}

// Extend the lock's TTL.
func (l *lock) SetTTL(ctx context.Context, ttl time.Duration) error {
	return l.l.Refresh(ctx, ttl, nil)
}

// Don't forget to release the lock when you're done.
func (l *lock) Unlock(ctx context.Context) error {
	return l.l.Release(ctx)
}
