package fsutil

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"mime"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gookit/goutil/fsutil"
)

// CreateFile creates a file and automatically creates a directory if the file directory does not exist.
func CreateFile(path string) (*os.File, error) {
	return fsutil.CreateFile(path, 0644, 0755)
}

// GenerateUploadPath builds an object key for uploaded files.
func GenerateUploadPath(userID int64, filename string) string {
	const defaultPrefix = "uploads"
	const emptyName = "file"

	filename = normalizeUploadFilename(filename)
	ext := sanitizeUploadExtension(path.Ext(filename))
	name := strings.TrimSuffix(filename, path.Ext(filename))
	name = sanitizeUploadName(name)
	if name == "" {
		name = emptyName
	}

	keyName := time.Now().Format("20060102150405") + "-" + randomUploadSuffix() + "-" + name
	if ext != "" {
		keyName += "." + ext
	}

	return path.Join(defaultPrefix, strconv.FormatInt(userID, 10), keyName)
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

// ParseMIMEType validates v as a bare MIME type and returns its normalized form.
func ParseMIMEType(v string) (string, error) {
	if v == "" {
		return "", errors.New("mime type is empty")
	}
	if strings.TrimSpace(v) != v {
		return "", errors.New("mime type must not contain leading or trailing spaces")
	}

	mediaType, _, err := mime.ParseMediaType(v)
	if err != nil {
		return "", err
	}

	typePart, subType, ok := strings.Cut(mediaType, "/")
	if !ok || typePart == "" || subType == "" || strings.Contains(subType, "/") {
		return "", errors.New("mime type must be in type/subtype form")
	}

	return mediaType, nil
}

func normalizeUploadFilename(filename string) string {
	filename = strings.ReplaceAll(filename, "\\", "/")
	filename = path.Base(filename)
	if filename == "." || filename == "/" {
		return ""
	}
	return filename
}

func sanitizeUploadName(name string) string {
	var b strings.Builder
	lastDash := false

	for _, r := range name {
		if isASCIIAlphaNumeric(r) {
			if 'A' <= r && r <= 'Z' {
				r = r + ('a' - 'A')
			}
			b.WriteRune(r)
			lastDash = false
			continue
		}

		if !lastDash && b.Len() > 0 {
			b.WriteByte('-')
			lastDash = true
		}
	}

	return strings.Trim(b.String(), "-")
}

func sanitizeUploadExtension(ext string) string {
	ext = strings.TrimPrefix(strings.ToLower(ext), ".")
	if ext == "" {
		return ""
	}

	var b strings.Builder
	for _, r := range ext {
		if isASCIIAlphaNumeric(r) {
			if 'A' <= r && r <= 'Z' {
				r = r + ('a' - 'A')
			}
			b.WriteRune(r)
		}
	}

	return b.String()
}

func isASCIIAlphaNumeric(r rune) bool {
	return ('a' <= r && r <= 'z') || ('A' <= r && r <= 'Z') || ('0' <= r && r <= '9')
}

func randomUploadSuffix() string {
	var buf [2]byte
	if _, err := rand.Read(buf[:]); err != nil {
		return "0000"
	}
	return hex.EncodeToString(buf[:])
}
