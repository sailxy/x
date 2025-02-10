package id

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewNanoID(t *testing.T) {
	id, err := NewNanoID(10)
	assert.NoError(t, err)
	assert.Len(t, id, 10)

	id2, err := NewNanoID(20)
	assert.NoError(t, err)
	assert.Len(t, id2, 20)
}
