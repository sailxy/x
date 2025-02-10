package redis

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	rds := New(Config{
		Addr: "localhost:6379",
	})
	s, err := rds.Ping(context.Background()).Result()
	assert.NoError(t, err)
	t.Log(s)
}
