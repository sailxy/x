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
