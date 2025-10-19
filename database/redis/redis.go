package redis

import "github.com/redis/go-redis/v9"

type Client = redis.Client

type Config struct {
	Addr     string
	Password string
}

func New(c Config) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     c.Addr,
		Password: c.Password,
	})
}
