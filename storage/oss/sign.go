package oss

import (
	"fmt"

	aliyunoss "github.com/aliyun/aliyun-oss-go-sdk/oss"
)

func (c *Client) SignURL(key string, cfg SignURLConfig) (*signURLResp, error) {
	opts := []aliyunoss.Option{}
	if cfg.ContentType != "" {
		opts = append(opts, aliyunoss.ContentType(cfg.ContentType))
	}
	if cfg.Callback != "" {
		opts = append(opts, aliyunoss.Callback(cfg.Callback))
	}
	if cfg.CallbackVar != "" {
		opts = append(opts, aliyunoss.CallbackVar(cfg.CallbackVar))
	}

	expiredInSec := cfg.ExpiredInSec
	if expiredInSec == 0 {
		expiredInSec = defaultExpiredInSec
	}

	method, err := parseHTTPMethod(cfg.HTTPMethod)
	if err != nil {
		return nil, err
	}

	signedURL, err := c.bucket.SignURL(key, method, expiredInSec, opts...)
	if err != nil {
		return nil, err
	}

	return &signURLResp{
		SignedURL:    signedURL,
		ExpiredInSec: expiredInSec,
	}, nil
}

func parseHTTPMethod(method HTTPMethod) (aliyunoss.HTTPMethod, error) {
	switch method {
	case "", HTTPPut:
		return aliyunoss.HTTPPut, nil
	case HTTPGet:
		return aliyunoss.HTTPGet, nil
	default:
		return "", fmt.Errorf("unsupported HTTPMethod %q, only GET and PUT are supported", method)
	}
}
