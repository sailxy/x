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

func TestParseMIMEType(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		want    string
		wantErr string
	}{
		{
			name:  "valid image jpeg",
			input: "image/jpeg",
			want:  "image/jpeg",
		},
		{
			name:  "valid application json",
			input: "application/json",
			want:  "application/json",
		},
		{
			name:  "normalizes uppercase",
			input: "TEXT/PLAIN",
			want:  "text/plain",
		},
		{
			name:    "empty string",
			input:   "",
			wantErr: "mime type is empty",
		},
		{
			name:    "missing slash",
			input:   "abc",
			wantErr: "type/subtype",
		},
		{
			name:    "missing subtype",
			input:   "image",
			wantErr: "type/subtype",
		},
		{
			name:    "extra slash",
			input:   "image//jpeg",
			wantErr: "mime:",
		},
		{
			name:    "leading and trailing spaces",
			input:   " image/jpeg ",
			wantErr: "leading or trailing spaces",
		},
		{
			name:    "invalid characters",
			input:   "image/jp<eg",
			wantErr: "mime:",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := ParseMIMEType(tt.input)
			if tt.wantErr != "" {
				assert.Error(t, err)
				assert.ErrorContains(t, err, tt.wantErr)
				assert.Empty(t, got)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
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
