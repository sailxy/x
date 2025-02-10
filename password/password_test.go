package password

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEncrypt(t *testing.T) {
	password := "123456"
	hash, err := Encrypt(password)
	assert.NoError(t, err)
	assert.Equal(t, "$2a", hash[:3])
	t.Log(hash)
}

func TestCheck(t *testing.T) {
	password := "123456"
	hash, err := Encrypt(password)
	assert.NoError(t, err)
	assert.NoError(t, Check(hash, password))
}
