package lock

import (
	"context"
	"fmt"
	"time"

	"github.com/bsm/redislock"
	"github.com/redis/go-redis/v9"
)

type Client struct {
	client *redis.Client
}

func NewClient(client *redis.Client) (*Client, error) {
	return &Client{
		client: client,
	}, nil
}

func (c *Client) NewLocker() *locker {
	return &locker{
		client: redislock.New(c.client),
	}
}

type Options struct{}

type locker struct {
	client *redislock.Client
	l      *redislock.Lock
}

func (l *locker) Lock(ctx context.Context, key string, ttl time.Duration, opts *Options) error {
	lock, err := l.client.Obtain(ctx, key, ttl, nil)
	if err != nil {
		return fmt.Errorf("failed to obtain lock: %w", err)
	}
	l.l = lock
	return nil
}

func (l *locker) TTL(ctx context.Context) (time.Duration, error) {
	return l.l.TTL(ctx)
}

func (l *locker) SetTTL(ctx context.Context, ttl time.Duration) error {
	return l.l.Refresh(ctx, ttl, nil)
}

func (l *locker) Unlock(ctx context.Context) error {
	return l.l.Release(ctx)
}
