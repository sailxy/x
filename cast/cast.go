package cast

import "github.com/spf13/cast"

func ToBool(v any) bool {
	return cast.ToBool(v)
}

func ToString(v any) string {
	return cast.ToString(v)
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
