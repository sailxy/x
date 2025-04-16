package cast

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type testStruct struct {
	Name string `json:"name"`
}

func TestToBool(t *testing.T) {
	r := ToBool(1)
	assert.True(t, r)
}

func TestToString(t *testing.T) {
	t.Run("test int", func(t *testing.T) {
		r := ToString(123456)
		assert.Equal(t, "123456", r)
	})
}

func TestToJSONString(t *testing.T) {
	r, err := ToJSONString(testStruct{Name: "test"})
	assert.NoError(t, err)
	assert.Equal(t, `{"name":"test"}`, r)
}

func TestToInt(t *testing.T) {
	r := ToInt("-123456")
	assert.Equal(t, -123456, r)
}

func TestToUInt(t *testing.T) {
	r := ToUInt("123456")
	assert.Equal(t, uint(123456), r)
}

func TestToInt64(t *testing.T) {
	r := ToInt64("-1444784865584")
	assert.Equal(t, int64(-1444784865584), r)
}

func TestToUInt64(t *testing.T) {
	r := ToUInt64("1444784865584")
	assert.Equal(t, uint64(1444784865584), r)
}
