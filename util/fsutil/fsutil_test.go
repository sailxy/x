package fsutil

import (
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateFile(t *testing.T) {
	path := "/tmp/tmpfile"
	f, err := CreateFile(path)
	assert.NoError(t, err)
	defer func() { _ = os.RemoveAll(path) }()
	t.Log(f.Name())
}

func TestFindModuleRoot(t *testing.T) {
	dir, err := os.Getwd()
	assert.NoError(t, err)

	roots := FindModuleRoot(dir)
	assert.NotEmpty(t, roots)
}

func TestGenerateUploadKey(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		userID   int64
		filename string
		pattern  string
	}{
		{
			name:     "normal filename",
			userID:   123,
			filename: "abc.png",
			pattern:  `^uploads/123/\d{14}-[0-9a-f]{4}-abc\.png$`,
		},
		{
			name:     "symbols become dashes",
			userID:   123,
			filename: "a_b c@d.png",
			pattern:  `^uploads/123/\d{14}-[0-9a-f]{4}-a-b-c-d\.png$`,
		},
		{
			name:     "empty stem falls back to file",
			userID:   123,
			filename: "___.PNG",
			pattern:  `^uploads/123/\d{14}-[0-9a-f]{4}-file\.png$`,
		},
		{
			name:     "repeated junk collapses",
			userID:   123,
			filename: "a---___***b.png",
			pattern:  `^uploads/123/\d{14}-[0-9a-f]{4}-a-b\.png$`,
		},
		{
			name:     "no extension",
			userID:   456,
			filename: "report",
			pattern:  `^uploads/456/\d{14}-[0-9a-f]{4}-report$`,
		},
		{
			name:     "drops path parts",
			userID:   789,
			filename: `foo/bar\\baz?.PDF`,
			pattern:  `^uploads/789/\d{14}-[0-9a-f]{4}-baz\.pdf$`,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			key := GenerateUploadPath(tt.userID, tt.filename)
			assert.Regexp(t, regexp.MustCompile(tt.pattern), key)
		})
	}
}

func TestGenerateUploadKeyRandomSuffix(t *testing.T) {
	t.Parallel()

	keys := make(map[string]struct{})
	for range 20 {
		key := GenerateUploadPath(123, "abc.png")
		parts := strings.Split(key, "/")
		if assert.Len(t, parts, 3) {
			nameParts := strings.Split(parts[2], "-")
			if assert.GreaterOrEqual(t, len(nameParts), 3) {
				assert.Regexp(t, regexp.MustCompile(`^[0-9a-f]{4}$`), nameParts[1])
			}
		}
		keys[key] = struct{}{}
	}

	assert.Greater(t, len(keys), 1)
}
