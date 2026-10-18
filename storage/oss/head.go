package oss

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	aliyunoss "github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss"
)

// HeadObject reads metadata only from the configured bucket using caller context.
func (c *Client) HeadObject(ctx context.Context, key string) (*ObjectMeta, error) {
	if key == "" {
		return nil, fmt.Errorf("OSS object key is empty")
	}
	result, err := c.sdkClient.HeadObject(ctx, &aliyunoss.HeadObjectRequest{Bucket: aliyunoss.Ptr(c.bucketName), Key: aliyunoss.Ptr(key)}, validateHeadSize)
	if err != nil {
		return nil, fmt.Errorf("head OSS object: %w", err)
	}
	return &ObjectMeta{Size: result.ContentLength, ContentType: aliyunoss.ToString(result.ContentType)}, nil
}

func validateHeadSize(options *aliyunoss.Options) {
	options.ResponseHandlers = append(options.ResponseHandlers, func(response *http.Response) error {
		if response.StatusCode < 200 || response.StatusCode >= 300 {
			return nil
		}
		if response.Request == nil {
			return fmt.Errorf("OSS response is missing request metadata")
		}
		if response.Request.Method != http.MethodHead {
			return nil
		}
		// V2 otherwise treats a missing size as zero, including in its Downloader.
		size, err := strconv.ParseInt(response.Header.Get("Content-Length"), 10, 64)
		if err != nil {
			return fmt.Errorf("parse OSS object size: %w", err)
		}
		if size < 0 {
			return fmt.Errorf("OSS object size is negative")
		}
		return nil
	})
}
