package oss

import (
	"context"
	"fmt"
	"time"

	aliyunoss "github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss"
)

func (c *Client) SignURL(key string, cfg SignURLConfig) (*signURLResp, error) {
	expires := cfg.ExpiredInSec
	if expires == 0 {
		expires = defaultExpiredInSec
	}
	if expires < 0 || expires > int64((1<<63-1)/time.Second) {
		return nil, fmt.Errorf("OSS presign expiration is invalid")
	}
	method, err := parseHTTPMethod(cfg.HTTPMethod)
	if err != nil {
		return nil, err
	}
	headers := map[string]string{}
	if cfg.ContentType != "" {
		headers["Content-Type"] = cfg.ContentType
	}
	if cfg.Callback != "" {
		headers["x-oss-callback"] = cfg.Callback
	}
	if cfg.CallbackVar != "" {
		headers["x-oss-callback-var"] = cfg.CallbackVar
	}
	common := aliyunoss.RequestCommon{Headers: headers}
	var request any
	if method == HTTPGet {
		get := &aliyunoss.GetObjectRequest{Bucket: aliyunoss.Ptr(c.bucketName), Key: aliyunoss.Ptr(key), RequestCommon: common}
		if cfg.ResponseContentDisposition != "" {
			get.ResponseContentDisposition = aliyunoss.Ptr(cfg.ResponseContentDisposition)
		}
		request = get
	} else {
		put := &aliyunoss.PutObjectRequest{Bucket: aliyunoss.Ptr(c.bucketName), Key: aliyunoss.Ptr(key), RequestCommon: common}
		if cfg.ResponseContentDisposition != "" {
			put.AddParameter("response-content-disposition", cfg.ResponseContentDisposition)
		}
		request = put
	}
	result, err := c.sdkClient.Presign(context.Background(), request, aliyunoss.PresignExpires(time.Duration(expires)*time.Second))
	if err != nil {
		return nil, fmt.Errorf("presign OSS URL: %w", err)
	}
	return &signURLResp{SignedURL: result.URL, ExpiredInSec: expires, SignedHeaders: result.SignedHeaders}, nil
}
func parseHTTPMethod(method HTTPMethod) (HTTPMethod, error) {
	switch method {
	case "", HTTPPut:
		return HTTPPut, nil
	case HTTPGet:
		return HTTPGet, nil
	default:
		return "", fmt.Errorf("unsupported HTTPMethod %q, only GET and PUT are supported", method)
	}
}
