package oss

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"time"

	aliyunoss "github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss"
	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss/credentials"
)

func New(c Config) (*Client, error) {
	endpoint := normalizeEndpoint(c.Endpoint)
	parsed, err := url.Parse(endpoint)
	if err != nil || parsed.Host == "" || (parsed.Scheme != "https" && parsed.Scheme != "http") || !regexp.MustCompile(`^[a-z0-9][a-z0-9-]{1,61}[a-z0-9]$`).MatchString(c.BucketName) {
		return nil, fmt.Errorf("OSS endpoint or bucket is invalid")
	}
	cfg := aliyunoss.LoadDefaultConfig().WithEndpoint(endpoint).WithCredentialsProvider(credentials.NewStaticCredentialsProvider(c.AccessKeyID, c.AccessKeySecret))
	if c.Region == "" {
		// Existing callers without Region retain V1 request/presign behavior.
		cfg.WithSignatureVersion(aliyunoss.SignatureVersionV1)
	} else {
		cfg.WithRegion(c.Region).WithConnectTimeout(5 * time.Second).WithReadWriteTimeout(10 * time.Second)
	}
	return &Client{region: c.Region, sdkClient: aliyunoss.NewClient(cfg), endpoint: endpoint, bucketName: c.BucketName, host: bucketHost(c.BucketName, endpoint), accessKeyID: c.AccessKeyID, accessKeySecret: c.AccessKeySecret}, nil
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

// ObjectBaseURL returns a stable object URL root, without read authorization.
func (c *Client) ObjectBaseURL() string { return strings.TrimRight(c.host, "/") + "/" }
