package printer

import (
	"github.com/gookit/color"
	"github.com/k0kubun/pp/v3"
)

func Print(s ...any) {
	_, _ = pp.Println(s...)
}

func Error(s ...any) {
	color.Error.Println(s...)
}
