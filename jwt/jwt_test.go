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
func TestParse(t *testing.T) {
	j := New(Config{
		Secret: []byte("123456"),
	})

	tests := []struct {
		name    string
		token   string
		wantErr bool
	}{
		{
			name:    "invalid token format",
			token:   "invalid.token.format",
			wantErr: true,
		},
		{
			name:    "empty token",
			token:   "",
			wantErr: true,
		},
		{
			name:    "token with wrong signature",
			token:   "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0.wrong-signature",
			wantErr: true,
		},
		{
			name: "valid token",
			token: func() string {
				token, _ := j.NewWithClaims(map[string]any{"test": "data"})
				t.Log("fk", token)
				return token
			}(),
			wantErr: false,
		},
		{
			name: "token with bearer",
			token: func() string {
				token, _ := j.NewWithClaims(map[string]any{"test": "data"})
				return "Bearer " + token
			}(),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims, err := j.Parse(tt.token)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, claims)
			}
		})
	}
}
