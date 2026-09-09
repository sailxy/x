package oss

import (
	"io"

	aliyunoss "github.com/aliyun/aliyun-oss-go-sdk/oss"
)

const defaultExpiredInSec int64 = 300

type HTTPMethod = aliyunoss.HTTPMethod

const (
	HTTPGet HTTPMethod = aliyunoss.HTTPGet
	HTTPPut HTTPMethod = aliyunoss.HTTPPut
)

type Config struct {
	Endpoint        string
	BucketName      string
	AccessKeyID     string
	AccessKeySecret string
}

type Client struct {
	endpoint        string
	bucketName      string
	host            string
	accessKeyID     string
	accessKeySecret string
	bucket          downloadBucketAPI
}

type downloadBucketAPI interface {
	SignURL(objectKey string, method aliyunoss.HTTPMethod, expiredInSec int64, options ...aliyunoss.Option) (string, error)
	GetObject(objectKey string, options ...aliyunoss.Option) (io.ReadCloser, error)
	GetObjectToFile(objectKey, filePath string, options ...aliyunoss.Option) error
}

type signURLResp struct {
	SignedURL    string
	ExpiredInSec int64
}

type SignURLConfig struct {
	ContentType  string
	HTTPMethod   HTTPMethod
	ExpiredInSec int64 // expired in seconds, default 300s
	Callback     string
	CallbackVar  string
}
