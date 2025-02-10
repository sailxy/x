package fsutil

import (
	"os"
	"path/filepath"

	"github.com/gookit/goutil/fsutil"
)

// CreateFile creates a file and automatically creates a directory if the file directory does not exist.
func CreateFile(path string) (*os.File, error) {
	return fsutil.CreateFile(path, 0644, 0755)
}

// https://github.com/golang/go/blob/9e3b1d53a012e98cfd02de2de8b1bd53522464d4/src/cmd/go/internal/modload/init.go#L1504C1-L1522C2
func FindModuleRoot(dir string) string {
	if dir == "" {
		return ""
	}
	dir = filepath.Clean(dir)

	// Look for enclosing go.mod.
	for {
		if fi, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil && !fi.IsDir() {
			return dir
		}
		d := filepath.Dir(dir)
		if d == dir {
			break
		}
		dir = d
	}
	return ""
}
