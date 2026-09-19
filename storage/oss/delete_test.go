package oss

import (
	"context"
	"testing"

	aliyunoss "github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/stretchr/testify/assert"
)

func (b *fakeBucket) DeleteObjects(objectKeys []string, options ...aliyunoss.Option) (aliyunoss.DeleteObjectsResult, error) {
	b.deleteObjectsKeys = objectKeys
	b.deleteObjectsOptions = len(options)
	return aliyunoss.DeleteObjectsResult{DeletedObjects: objectKeys}, nil
}

func TestDeleteObjectsUsesProvidedContextAndReturnsDeletedKeys(t *testing.T) {
	bucket := &fakeBucket{}
	client := &Client{bucket: bucket}

	keys, err := client.DeleteObjects(context.Background(), []string{"path/to/a.txt", "path/to/b.txt"})

	assert.NoError(t, err)
	assert.Equal(t, []string{"path/to/a.txt", "path/to/b.txt"}, keys)
	assert.Equal(t, []string{"path/to/a.txt", "path/to/b.txt"}, bucket.deleteObjectsKeys)
	assert.Equal(t, 1, bucket.deleteObjectsOptions)
}

func TestDeleteObjectsSkipsSDKCallWithoutKeys(t *testing.T) {
	bucket := &fakeBucket{}
	client := &Client{bucket: bucket}

	keys, err := client.DeleteObjects(context.Background(), nil)

	assert.NoError(t, err)
	assert.Nil(t, keys)
	assert.Empty(t, bucket.deleteObjectsKeys)
}
