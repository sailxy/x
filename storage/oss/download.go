package oss

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"strings"

	aliyunoss "github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss"
)

func (c *Client) DownloadURL(rawURL string) (io.ReadCloser, error) {
	key, err := parseObjectKeyFromURL(rawURL)
	if err != nil {
		return nil, fmt.Errorf("parse object key from url: %w", err)
	}

	result, err := c.sdkClient.GetObject(context.Background(), &aliyunoss.GetObjectRequest{Bucket: aliyunoss.Ptr(c.bucketName), Key: aliyunoss.Ptr(key)})
	if err != nil {
		return nil, fmt.Errorf("download object %q: %w", key, err)
	}

	return result.Body, nil
}

func (c *Client) DownloadURLToFile(rawURL, filePath string) error {
	key, err := parseObjectKeyFromURL(rawURL)
	if err != nil {
		return fmt.Errorf("parse object key from url: %w", err)
	}

	if _, err := aliyunoss.NewDownloader(c.sdkClient, func(options *aliyunoss.DownloaderOptions) {
		options.ClientOptions = append(options.ClientOptions, validateHeadSize)
	}).DownloadFile(context.Background(), &aliyunoss.GetObjectRequest{Bucket: aliyunoss.Ptr(c.bucketName), Key: aliyunoss.Ptr(key)}, filePath); err != nil {
		return fmt.Errorf("download object %q to file: %w", key, err)
	}

	return nil
}

func parseObjectKeyFromURL(rawURL string) (string, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", fmt.Errorf("parse url: %w", err)
	}
	if u.Scheme == "" || u.Host == "" {
		return "", errors.New("url scheme and host are required")
	}

	escapedPath := strings.TrimPrefix(u.EscapedPath(), "/")
	if escapedPath == "" {
		return "", errors.New("object key is empty")
	}

	key, err := url.PathUnescape(escapedPath)
	if err != nil {
		return "", fmt.Errorf("decode object key: %w", err)
	}
	if key == "" {
		return "", errors.New("object key is empty")
	}

	return key, nil
}
