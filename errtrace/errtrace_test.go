package errtrace

import (
	"errors"
	"testing"
)

func TestWrap(t *testing.T) {
	err1 := errors.New("test error")
	err2 := Wrap(err1)
	Print(err2)
}
