package errtrace

import (
	"github.com/ztrue/tracerr"
)

func Wrap(err error) error {
	return tracerr.Wrap(err)
}

func Print(err error) {
	tracerr.Print(err)
}
