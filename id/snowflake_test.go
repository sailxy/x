package id

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewSnowflakeID(t *testing.T) {
	id, err := NewSnowflakeID()
	assert.NoError(t, err)
	t.Log(id)
}
