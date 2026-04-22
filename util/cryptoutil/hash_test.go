package cryptoutil

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMD5(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		want string
	}{
		{
			name: "hello",
			data: []byte("hello"),
			want: "5d41402abc4b2a76b9719d911017c592",
		},
		{
			name: "empty",
			data: []byte(""),
			want: "d41d8cd98f00b204e9800998ecf8427e",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, MD5(tt.data))
			assert.Equal(t, tt.want, MD5String(string(tt.data)))
		})
	}
}

func TestSHA256(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		want string
	}{
		{
			name: "hello",
			data: []byte("hello"),
			want: "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824",
		},
		{
			name: "empty",
			data: []byte(""),
			want: "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, SHA256(tt.data))
			assert.Equal(t, tt.want, SHA256String(string(tt.data)))
		})
	}
}

func TestStdlibReferenceValuesStayStable(t *testing.T) {
	md5Sum := md5.Sum([]byte("hello"))
	assert.Equal(t, "5d41402abc4b2a76b9719d911017c592", hex.EncodeToString(md5Sum[:]))

	sha256Sum := sha256.Sum256([]byte("hello"))
	assert.Equal(t, "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824", hex.EncodeToString(sha256Sum[:]))
}
