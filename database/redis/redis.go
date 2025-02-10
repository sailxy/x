package redis

import "github.com/redis/go-redis/v9"

type Config struct {
	Addr string
}

func New(c Config) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr: c.Addr,
	})
}
