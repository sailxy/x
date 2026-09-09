package oss

import (
	"encoding/base64"
	"encoding/json"
	"time"

	"github.com/sailxy/x/util/cryptoutil"
)

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
