package oss

import (
	"strings"

	aliyunoss "github.com/aliyun/aliyun-oss-go-sdk/oss"
)

func New(c Config) (*Client, error) {
	endpoint := normalizeEndpoint(c.Endpoint)
	client, err := aliyunoss.New(endpoint, c.AccessKeyID, c.AccessKeySecret)
	if err != nil {
		return nil, err
	}

	bucket, err := client.Bucket(c.BucketName)
	if err != nil {
		return nil, err
	}

	host := bucketHost(c.BucketName, endpoint)
	return &Client{
		endpoint:        endpoint,
		bucketName:      c.BucketName,
		host:            host,
		accessKeyID:     c.AccessKeyID,
		accessKeySecret: c.AccessKeySecret,
		bucket:          bucket,
	}, nil
}

func normalizeEndpoint(endpoint string) string {
	if strings.HasPrefix(endpoint, "http://") || strings.HasPrefix(endpoint, "https://") {
		return endpoint
	}
	return "https://" + endpoint
}

func bucketHost(bucketName, endpoint string) string {
	return strings.TrimRight(strings.Replace(endpoint, "://", "://"+bucketName+".", 1), "/")
}
