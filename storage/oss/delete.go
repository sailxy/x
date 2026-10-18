package oss

import (
	"context"
	"fmt"

	aliyunoss "github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss"
)

// DeleteObject removes an object with the supplied context.
func (c *Client) DeleteObject(ctx context.Context, key string) error {
	if _, err := c.sdkClient.DeleteObject(ctx, &aliyunoss.DeleteObjectRequest{Bucket: aliyunoss.Ptr(c.bucketName), Key: aliyunoss.Ptr(key)}); err != nil {
		return fmt.Errorf("delete object: %w", err)
	}
	return nil
}

// DeleteObjects returns only keys reported as deleted, including absent objects.
// Requests are limited to 1000 keys; callers must retry any missing returned key.
func (c *Client) DeleteObjects(ctx context.Context, keys []string) ([]string, error) {
	if len(keys) == 0 {
		return nil, nil
	}
	objects := make([]aliyunoss.ObjectIdentifier, 0, len(keys))
	for _, key := range keys {
		objects = append(objects, aliyunoss.ObjectIdentifier{Key: aliyunoss.Ptr(key)})
	}
	result, err := c.sdkClient.DeleteMultipleObjects(ctx, &aliyunoss.DeleteMultipleObjectsRequest{Bucket: aliyunoss.Ptr(c.bucketName), Delete: &aliyunoss.Delete{Objects: objects, Quiet: false}})
	if err != nil {
		return nil, fmt.Errorf("delete objects: %w", err)
	}
	deleted := make([]string, 0, len(result.DeletedObjects))
	for _, object := range result.DeletedObjects {
		deleted = append(deleted, aliyunoss.ToString(object.Key))
	}
	return deleted, nil
}
