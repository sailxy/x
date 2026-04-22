package cryptoutil

import (
	"crypto/hmac"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHMACSHA256(t *testing.T) {
	data := []byte("hello")
	secret := []byte("world")

	mac := hmac.New(sha256.New, secret)
	_, err := mac.Write(data)
	assert.NoError(t, err)

	want := hex.EncodeToString(mac.Sum(nil))

	assert.Equal(t, want, HMACSHA256(data, secret))
	assert.Equal(t, want, HMACSHA256String(string(data), string(secret)))
}

func TestHMACSHA1Base64(t *testing.T) {
	data := []byte("hello")
	secret := []byte("world")

	mac := hmac.New(sha1.New, secret)
	_, err := mac.Write(data)
	assert.NoError(t, err)

	want := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	assert.Equal(t, want, HMACSHA1Base64(data, secret))
	assert.Equal(t, want, HMACSHA1Base64String(string(data), string(secret)))
}

func TestVerifyHMACSHA256(t *testing.T) {
	signature := HMACSHA256([]byte("hello"), []byte("world"))

	tests := []struct {
		name      string
		data      []byte
		secret    []byte
		signature string
		want      bool
		wantErr   bool
	}{
		{
			name:      "valid signature",
			data:      []byte("hello"),
			secret:    []byte("world"),
			signature: signature,
			want:      true,
		},
		{
			name:      "modified payload",
			data:      []byte("hello!"),
			secret:    []byte("world"),
			signature: signature,
			want:      false,
		},
		{
			name:      "wrong secret",
			data:      []byte("hello"),
			secret:    []byte("world!"),
			signature: signature,
			want:      false,
		},
		{
			name:      "invalid hex signature",
			data:      []byte("hello"),
			secret:    []byte("world"),
			signature: "zz",
			wantErr:   true,
		},
		{
			name:      "empty payload and secret",
			data:      []byte(""),
			secret:    []byte(""),
			signature: HMACSHA256([]byte(""), []byte("")),
			want:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := VerifyHMACSHA256(tt.data, tt.secret, tt.signature)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)

			got, err = VerifyHMACSHA256String(string(tt.data), string(tt.secret), tt.signature)
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
