package oss

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDownloadURLRejectsInvalidURLBeforeSDKCall(t *testing.T) {
	c := &Client{}
	body, err := c.DownloadURL("https://bucket.example.test/")
	require.Nil(t, body)
	require.Error(t, err)
	require.Contains(t, err.Error(), "parse object key")
	require.Error(t, c.DownloadURLToFile("relative/path", "unused"))
}
func TestDownloadURLUsesParsedObjectKey(t *testing.T) {
	c := localOSSClient(t, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/private-test/path/file.txt", r.URL.Path)
		w.Header().Set("Content-Length", "6")
		_, err := fmt.Fprint(w, "object")
		require.NoError(t, err)
	})
	body, err := c.DownloadURL("https://different.example.test/path/file.txt?ignored=1")
	require.NoError(t, err)
	data, err := io.ReadAll(body)
	require.NoError(t, err)
	require.NoError(t, body.Close())
	require.Equal(t, "object", string(data))
}
func TestDownloadURLToFileUsesParsedObjectKey(t *testing.T) {
	c := localOSSClient(t, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/private-test/path/file.txt", r.URL.Path)
		w.Header().Set("Content-Length", "6")
		w.Header().Set("ETag", `"test-etag"`)
		if r.Method == http.MethodHead {
			return
		}
		if r.Header.Get("Range") != "" {
			w.Header().Set("Content-Range", "bytes 0-5/6")
			w.WriteHeader(http.StatusPartialContent)
		}
		_, err := fmt.Fprint(w, "object")
		require.NoError(t, err)
	})
	target := filepath.Join(t.TempDir(), "file.txt")
	require.NoError(t, c.DownloadURLToFile("https://bucket.example.test/path/file.txt?ignored=1", target))
	data, err := os.ReadFile(target)
	require.NoError(t, err)
	require.Equal(t, "object", string(data))
}
func TestDownloadURLToFileFailureKeepsExistingFile(t *testing.T) {
	c := localOSSClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodHead {
			w.Header().Set("Content-Length", "6")
			w.Header().Set("ETag", `"test-etag"`)
			return
		}
		w.WriteHeader(http.StatusForbidden)
		_, err := fmt.Fprint(w, `<Error><Code>AccessDenied</Code></Error>`)
		require.NoError(t, err)
	})
	target := filepath.Join(t.TempDir(), "file.txt")
	require.NoError(t, os.WriteFile(target, []byte("original"), 0600))
	require.Error(t, c.DownloadURLToFile("https://bucket.example.test/path/file.txt", target))
	data, err := os.ReadFile(target)
	require.NoError(t, err)
	require.Equal(t, "original", string(data))
}
func TestDownloadURLIntegration(t *testing.T) {
	requireAliyunOSSEnv(t)
	objectURL := os.Getenv("ALIYUN_OSS_DOWNLOAD_URL")
	if strings.TrimSpace(objectURL) == "" {
		t.Skip("ALIYUN_OSS_DOWNLOAD_URL is not set")
	}
	c, err := new()
	require.NoError(t, err)
	body, err := c.DownloadURL(objectURL)
	require.NoError(t, err)
	require.NoError(t, body.Close())
}

func TestDownloadURLToFileRejectsMissingSourceSize(t *testing.T) {
	c := localOSSClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodHead {
			return
		}
		w.Header().Set("Content-Length", "6")
		_, err := fmt.Fprint(w, "object")
		require.NoError(t, err)
	})
	target := filepath.Join(t.TempDir(), "file.txt")
	require.NoError(t, os.WriteFile(target, []byte("original"), 0600))
	require.Error(t, c.DownloadURLToFile("https://bucket.example.test/path/file.txt", target))
	data, err := os.ReadFile(target)
	require.NoError(t, err)
	require.Equal(t, "original", string(data))
}
