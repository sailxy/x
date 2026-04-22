package oss

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	aliyunoss "github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/sailxy/x/util/cryptoutil"
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
	bucket          *aliyunoss.Bucket
}

func New(c Config) (*Client, error) {
	client, err := aliyunoss.New(c.Endpoint, c.AccessKeyID, c.AccessKeySecret)
	if err != nil {
		return nil, err
	}

	bucket, err := client.Bucket(c.BucketName)
	if err != nil {
		return nil, err
	}

	host := "https://" + c.BucketName + "." + c.Endpoint
	return &Client{
		endpoint:        c.Endpoint,
		bucketName:      c.BucketName,
		host:            host,
		accessKeyID:     c.AccessKeyID,
		accessKeySecret: c.AccessKeySecret,
		bucket:          bucket,
	}, nil
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

type postConfig struct {
	Expiration string  `json:"expiration"`
	Conditions [][]any `json:"conditions"`
}

type policyToken struct {
	AccessKeyId string
	Host        string
	Expire      int64
	Signature   string
	Policy      string
	Directory   string
	Callback    string
}

func (c *Client) PostInfo(dir string) (*policyToken, error) {
	now := time.Now().Unix()
	expireEnd := now + defaultExpiredInSec
	tokenExpire := gmtISO8601(expireEnd)

	// Create post config json.
	var cfg postConfig
	cfg.Expiration = tokenExpire
	condition1 := []any{"starts-with", "$key", dir}
	condition2 := []any{"content-length-range", 1, 10485760}
	cfg.Conditions = append(cfg.Conditions, condition1)
	cfg.Conditions = append(cfg.Conditions, condition2)

	// Calucate signature.
	result, err := json.Marshal(cfg)
	if err != nil {
		return nil, err
	}
	debyte := base64.StdEncoding.EncodeToString(result)
	signedStr := signPostPolicy(debyte, c.accessKeySecret)

	var policyToken policyToken
	policyToken.AccessKeyId = c.accessKeyID
	policyToken.Host = c.host
	policyToken.Expire = expireEnd
	policyToken.Signature = signedStr
	policyToken.Directory = dir
	policyToken.Policy = debyte
	return &policyToken, nil
}

func gmtISO8601(expireEnd int64) string {
	return time.Unix(expireEnd, 0).UTC().Format("2006-01-02T15:04:05Z")
}

func signPostPolicy(policy, secret string) string {
	return cryptoutil.HMACSHA1Base64String(policy, secret)
}
