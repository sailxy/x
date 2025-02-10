package jwt

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	claims := map[string]any{
		"id":  1,
		"iss": "issuer",
		"exp": time.Now().Add(time.Hour).Unix(),
		"data": map[string]any{
			"user_id":   1,
			"user_name": "name",
		},
	}
	j := New(Config{
		Secret: []byte("123456"),
	})
	token, err := j.NewWithClaims(claims)
	assert.NoError(t, err)
	t.Log(token)

	c, err := j.Parse(token)
	assert.NoError(t, err)
	t.Log(c)
}
