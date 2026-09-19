package oss

import (
	"context"
	"io"
	"os"
	"strings"
	"testing"

	aliyunoss "github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/stretchr/testify/assert"
)

type fakeBucket struct {
	getObjectKey    string
	deleteObjectKey string
	deleteOptions   int
	getObjectToFile struct {
		key      string
		filePath string
	}
}

func (b *fakeBucket) DeleteObject(key string, options ...aliyunoss.Option) error {
	b.deleteObjectKey = key
	b.deleteOptions = len(options)
	return nil
}

func TestDeleteObjectUsesProvidedContext(t *testing.T) {
	bucket := &fakeBucket{}
	client := &Client{bucket: bucket}

	err := client.DeleteObject(context.Background(), "path/to/file.txt")

	assert.NoError(t, err)
	assert.Equal(t, "path/to/file.txt", bucket.deleteObjectKey)
	assert.Equal(t, 1, bucket.deleteOptions)
}

func (b *fakeBucket) SignURL(string, HTTPMethod, int64, ...aliyunoss.Option) (string, error) {
	return "", nil
}

func (b *fakeBucket) GetObject(key string, _ ...aliyunoss.Option) (io.ReadCloser, error) {
	b.getObjectKey = key
	return io.NopCloser(strings.NewReader("object")), nil
}

func (b *fakeBucket) GetObjectToFile(key, filePath string, _ ...aliyunoss.Option) error {
	b.getObjectToFile.key = key
	b.getObjectToFile.filePath = filePath
	return nil
}

func TestDownloadURLRejectsInvalidURLBeforeSDKCall(t *testing.T) {
	client := &Client{}

	rc, err := client.DownloadURL("https://bucket.oss-cn-hangzhou.aliyuncs.com/")

	assert.Nil(t, rc)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "parse object key from url")
}

func TestDownloadURLToFileRejectsInvalidURLBeforeSDKCall(t *testing.T) {
	client := &Client{}

	err := client.DownloadURLToFile("https://bucket.oss-cn-hangzhou.aliyuncs.com/", "/tmp/unused")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "parse object key from url")
}

func TestDownloadURLUsesParsedObjectKey(t *testing.T) {
	bucket := &fakeBucket{}
	client := &Client{bucket: bucket}

	rc, err := client.DownloadURL("https://bucket.oss-cn-hangzhou.aliyuncs.com/path/to/file.txt?Expires=1")
	if assert.NoError(t, err) {
		assert.NoError(t, rc.Close())
	}

	assert.Equal(t, "path/to/file.txt", bucket.getObjectKey)
}

func TestDownloadURLToFileUsesParsedObjectKey(t *testing.T) {
	bucket := &fakeBucket{}
	client := &Client{bucket: bucket}

	err := client.DownloadURLToFile("https://bucket.oss-cn-hangzhou.aliyuncs.com/path/to/file.txt?Expires=1", "/tmp/file.txt")

	assert.NoError(t, err)
	assert.Equal(t, "path/to/file.txt", bucket.getObjectToFile.key)
	assert.Equal(t, "/tmp/file.txt", bucket.getObjectToFile.filePath)
}

func TestDownloadURLIntegration(t *testing.T) {
	requireAliyunOSSEnv(t)

	objectURL := os.Getenv("ALIYUN_OSS_DOWNLOAD_URL")
	if strings.TrimSpace(objectURL) == "" {
		t.Skip("ALIYUN_OSS_DOWNLOAD_URL is not set")
	}

	client, err := new()
	if !assert.NoError(t, err) {
		return
	}

	rc, err := client.DownloadURL(objectURL)
	if assert.NoError(t, err) {
		assert.NoError(t, rc.Close())
	}
}
