package oss

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"hash"
	"io"
	"time"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

const expiredInSec int64 = 300

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
	bucket          *oss.Bucket
}

func New(c Config) (*Client, error) {
	client, err := oss.New(c.Endpoint, c.AccessKeyID, c.AccessKeySecret)
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
	SignedURL string
}

type SignURLConfig struct {
	ContentType string
	Callback    string
	CallbackVar string
}

func (c *Client) SignURL(key string, cfg SignURLConfig) (*signURLResp, error) {
	opts := []oss.Option{}
	if cfg.ContentType != "" {
		opts = append(opts, oss.ContentType(cfg.ContentType))
	}
	if cfg.Callback != "" {
		opts = append(opts, oss.Callback(cfg.Callback))
	}
	if cfg.CallbackVar != "" {
		opts = append(opts, oss.CallbackVar(cfg.CallbackVar))
	}

	signedURL, err := c.bucket.SignURL(key, oss.HTTPPut, expiredInSec, opts...)
	if err != nil {
		return nil, err
	}

	return &signURLResp{
		SignedURL: signedURL,
	}, nil
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
	expireEnd := now + expiredInSec
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
	h := hmac.New(func() hash.Hash { return sha1.New() }, []byte(c.accessKeySecret))
	_, err = io.WriteString(h, debyte)
	if err != nil {
		return nil, err
	}
	signedStr := base64.StdEncoding.EncodeToString(h.Sum(nil))

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
