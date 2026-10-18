package oss

import (
	"time"

	aliyunoss "github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss"
)

const defaultExpiredInSec int64 = 300

type HTTPMethod string

const (
	HTTPGet HTTPMethod = "GET"
	HTTPPut HTTPMethod = "PUT"
)

type Config struct {
	Region          string
	Endpoint        string
	BucketName      string
	AccessKeyID     string
	AccessKeySecret string
}

type Client struct {
	region          string
	sdkClient       *aliyunoss.Client
	endpoint        string
	bucketName      string
	host            string
	accessKeyID     string
	accessKeySecret string
}

type signURLResp struct {
	// SignedHeaders must accompany the request, especially for V4 uploads.
	SignedHeaders map[string]string
	SignedURL     string
	ExpiredInSec  int64
}

type SignURLConfig struct {
	ContentType                string
	HTTPMethod                 HTTPMethod
	ExpiredInSec               int64 // expired in seconds, default 300s
	Callback                   string
	CallbackVar                string
	ResponseContentDisposition string // Optional download response header, included in the URL signature.
}

// PostPolicyConfig controls the signed restrictions of a V4 browser POST.
type PostPolicyConfig struct {
	ExpiredInSec    int64
	MaxFileSize     int64
	ForbidOverwrite bool
	ObjectACL       string
}

// PostPolicy contains browser form fields, without the access key secret.
type PostPolicy struct {
	URL       string
	Fields    map[string]string
	ExpiresAt time.Time
}

// BucketState contains the configured bucket's ACL and versioning status.
type BucketState struct {
	ACL              string
	VersioningStatus string
}

// ObjectMeta is the size and content type returned by a configured bucket HEAD.
type ObjectMeta struct {
	Size        int64
	ContentType string
}
