package printer

import (
	"errors"
	"testing"
)

type testStruct struct {
	Name string
}

func TestPrint(t *testing.T) {
	Print("print", "test", 123456)
	Print("print struct", testStruct{Name: "test"})
	Print("print map", map[string]string{"foo": "bar", "hello": "world"})
	Print("print error", errors.New("test error"))
	Print("print slice", []int{1, 2, 3})
}

func TestError(t *testing.T) {
	Error("error", "test", 123456)
	Error("error struct", testStruct{Name: "test"})
	Error("error map", map[string]string{"foo": "bar", "hello": "world"})
	Error("error error", errors.New("test error"))
	Error("error slice", []int{1, 2, 3})
}
