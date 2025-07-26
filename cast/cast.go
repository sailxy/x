package cast

import (
	"encoding/json"

	"github.com/spf13/cast"
)

func ToBool(v any) bool {
	return cast.ToBool(v)
}

func ToString(v any) string {
	return cast.ToString(v)
}

// ToJSONString converts any struct to JSON string.
func ToJSONString(v any) (string, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func ToInt(v any) int {
	return cast.ToInt(v)
}

func ToUInt(v any) uint {
	return cast.ToUint(v)
}

func ToInt64(v any) int64 {
	return cast.ToInt64(v)
}

func ToUInt64(v any) uint64 {
	return cast.ToUint64(v)
}

func ToIntSlice(v any) []int {
	return cast.ToIntSlice(v)
}
