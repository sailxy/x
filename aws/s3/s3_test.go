package s3

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPresignPutObject(t *testing.T) {
	client := New(Config{
		Bucket: "bucket",
	})
	presign, err := client.PresignPutObject(context.Background(), "path/to/filename.txt")
	if assert.NoError(t, err) {
		t.Log(presign.URL)
	}
}

func TestPutObject(t *testing.T) {
	client := New(Config{
		Bucket: "bucket",
	})
	obj, err := client.PutObject(context.Background(), "path/to/filename.txt", []byte("hello world"))
	if assert.NoError(t, err) {
		t.Log(obj.Key, obj.ETag)
	}
}
