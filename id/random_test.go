package id

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewRandomNumber(t *testing.T) {
	num, err := NewRandomNumber(6)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, num, 100000)
	assert.LessOrEqual(t, num, 999999)
}
