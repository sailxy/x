package oss

import (
	"context"
	"fmt"

	aliyunoss "github.com/aliyun/aliyun-oss-go-sdk/oss"
)

// DeleteObject removes an object with the supplied context.
func (c *Client) DeleteObject(ctx context.Context, objectKey string) error {
	if err := c.bucket.DeleteObject(objectKey, aliyunoss.WithContext(ctx)); err != nil {
		return fmt.Errorf("delete object %q: %w", objectKey, err)
	}
	return nil
}

// DeleteObjects removes multiple objects in one request (at most 1000 keys per
// OSS call) with the supplied context and returns the keys reported as deleted;
// keys that do not exist count as deleted, so any requested key missing from the
// result failed to delete.
func (c *Client) DeleteObjects(ctx context.Context, objectKeys []string) ([]string, error) {
	if len(objectKeys) == 0 {
		return nil, nil
	}
	result, err := c.bucket.DeleteObjects(objectKeys, aliyunoss.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("delete objects: %w", err)
	}
	return result.DeletedObjects, nil
}
