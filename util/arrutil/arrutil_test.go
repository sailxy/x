package arrutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContains(t *testing.T) {
	res := Contains([]string{"a", "b", "c"}, "a")
	assert.True(t, res)
}

func TestUnique(t *testing.T) {
	tests := []struct {
		name string
		arr  []string
		want []string
	}{
		{
			name: "removes duplicates and keeps first occurrence order",
			arr:  []string{"a", "b", "a", "c", "b"},
			want: []string{"a", "b", "c"},
		},
		{
			name: "keeps unique values unchanged",
			arr:  []string{"a", "b", "c"},
			want: []string{"a", "b", "c"},
		},
		{
			name: "handles empty slice",
			arr:  []string{},
			want: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, Unique(tt.arr))
		})
	}
}
