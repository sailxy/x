package id

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewXID(t *testing.T) {
	xid, err := NewXID()
	assert.NoError(t, err)
	t.Log(xid)
}
