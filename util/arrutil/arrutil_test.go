package arrutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContains(t *testing.T) {
	res := Contains([]string{"a", "b", "c"}, "a")
	assert.True(t, res)
}
