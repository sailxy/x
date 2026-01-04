package id

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewUUID(t *testing.T) {
	uuid, err := NewUUID()
	assert.NoError(t, err)
	t.Log(uuid)
}
