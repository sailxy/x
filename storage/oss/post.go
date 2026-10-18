package oss

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"regexp"
	"time"
)

// SignPostPolicy signs a V4 browser POST with explicit expiry and restrictions.
func (c *Client) SignPostPolicy(dir string, cfg PostPolicyConfig) (*PostPolicy, error) {
	return c.signPostPolicyV4(dir, cfg, time.Now())
}

func (c *Client) signPostPolicyV4(dir string, cfg PostPolicyConfig, now time.Time) (*PostPolicy, error) {
	if dir == "" || cfg.ExpiredInSec <= 0 || cfg.ExpiredInSec > int64((1<<63-1)/time.Second) || cfg.MaxFileSize <= 0 ||
		!regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)+$`).MatchString(c.region) || c.accessKeyID == "" || c.accessKeySecret == "" {
		return nil, fmt.Errorf("OSS V4 POST configuration is invalid")
	}
	switch cfg.ObjectACL {
	case "", "private", "public-read", "public-read-write":
	default:
		return nil, fmt.Errorf("OSS POST object ACL is invalid")
	}
	now = now.UTC().Truncate(time.Second)
	expiresAt := now.Add(time.Duration(cfg.ExpiredInSec) * time.Second)
	credential := c.accessKeyID + "/" + now.Format("20060102") + "/" + c.region + "/oss/aliyun_v4_request"
	fields := map[string]string{"x-oss-signature-version": "OSS4-HMAC-SHA256", "x-oss-credential": credential, "x-oss-date": now.Format("20060102T150405Z"), "success_action_status": "204"}
	if cfg.ForbidOverwrite {
		fields["x-oss-forbid-overwrite"] = "true"
	}
	if cfg.ObjectACL != "" {
		fields["x-oss-object-acl"] = cfg.ObjectACL
	}
	conditions := []any{map[string]string{"bucket": c.bucketName}, []any{"starts-with", "$key", dir}, []any{"content-length-range", 0, cfg.MaxFileSize}}
	for _, key := range []string{"x-oss-signature-version", "x-oss-credential", "x-oss-date", "x-oss-forbid-overwrite", "x-oss-object-acl", "success_action_status"} {
		if value, ok := fields[key]; ok {
			conditions = append(conditions, map[string]string{key: value})
		}
	}
	policy, err := json.Marshal(struct {
		Expiration string `json:"expiration"`
		Conditions []any  `json:"conditions"`
	}{expiresAt.Format("2006-01-02T15:04:05.000Z"), conditions})
	if err != nil {
		return nil, fmt.Errorf("encode OSS POST policy: %w", err)
	}
	fields["policy"] = base64.StdEncoding.EncodeToString(policy)
	key := signV4([]byte("aliyun_v4"+c.accessKeySecret), now.Format("20060102"))
	for _, value := range []string{c.region, "oss", "aliyun_v4_request"} {
		key = signV4(key, value)
	}
	fields["x-oss-signature"] = hex.EncodeToString(signV4(key, fields["policy"]))
	return &PostPolicy{URL: c.ObjectBaseURL(), Fields: fields, ExpiresAt: expiresAt}, nil
}

func signV4(key []byte, value string) []byte {
	hash := hmac.New(sha256.New, key)
	// hash.Hash.Write never returns an error.
	if _, err := hash.Write([]byte(value)); err != nil {
		panic("HMAC-SHA256 write failed")
	}
	return hash.Sum(nil)
}
