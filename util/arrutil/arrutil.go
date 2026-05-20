package arrutil

import (
	"github.com/gookit/goutil/arrutil"
	"github.com/samber/lo"
)

// Contains checks if the given value is in the array.
func Contains(arr any, val any) bool {
	return arrutil.Contains(arr, val)
}

func Unique[T comparable](arr []T) []T {
	return lo.Uniq(arr)
}
