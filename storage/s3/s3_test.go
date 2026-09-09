package s3

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPresignPutObject(t *testing.T) {
	bucket := testBucket(t)
	client := New(Config{
		Bucket: bucket,
	})
	presign, err := client.PresignPutObject(context.Background(), "path/to/filename.txt")
	if assert.NoError(t, err) {
		t.Log(presign.URL)
	}
}

func TestPutObject(t *testing.T) {
	bucket := testBucket(t)
	client := New(Config{
		Bucket: bucket,
	})
	obj, err := client.PutObject(context.Background(), "path/to/filename.txt", []byte("hello world"))
	if assert.NoError(t, err) {
		t.Log(obj.Key, obj.ETag)
	}
}

func testBucket(t *testing.T) string {
	t.Helper()
	bucket := os.Getenv("AWS_S3_TEST_BUCKET")
	if bucket == "" {
		t.Skip("set AWS_S3_TEST_BUCKET to run S3 integration tests")
	}
	return bucket
}
