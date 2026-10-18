package oss

import (
	"encoding/base64"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestSignPostPolicyV4_RestrictionsAndIndependentVector(t *testing.T) {
	client, err := New(Config{Endpoint: "oss-cn-hangzhou.aliyuncs.com", BucketName: "private-test", Region: "cn-hangzhou", AccessKeyID: "test-ak", AccessKeySecret: "test-secret"})
	require.NoError(t, err)
	now := time.Date(2026, 10, 18, 12, 0, 0, 0, time.UTC)
	const prefix = "education/uploads/123/2026/10/18/1792324800-0123456789abcdef0123456789abcdef/"
	cfg := PostPolicyConfig{ExpiredInSec: 600, MaxFileSize: 20971520, ForbidOverwrite: true, ObjectACL: "private"}
	post, err := client.signPostPolicyV4(prefix, cfg, now)
	require.NoError(t, err)
	require.Equal(t, "https://private-test.oss-cn-hangzhou.aliyuncs.com/", post.URL)
	require.Equal(t, now.Add(10*time.Minute), post.ExpiresAt)
	// Independently calculated with Python HMAC-SHA256.
	require.Equal(t, "d0431ac716b6f84229308576f65b036c5d13965cfb5fe2a30090cfe185672317", post.Fields["x-oss-signature"])
	raw, err := base64.StdEncoding.DecodeString(post.Fields["policy"])
	require.NoError(t, err)
	for _, condition := range []string{`{"bucket":"private-test"}`, `["starts-with","$key","` + prefix + `"]`, `["content-length-range",0,20971520]`, `{"x-oss-forbid-overwrite":"true"}`, `{"x-oss-object-acl":"private"}`, `{"x-oss-date":"20261018T120000Z"}`, `{"x-oss-signature-version":"OSS4-HMAC-SHA256"}`, `{"x-oss-credential":"test-ak/20261018/cn-hangzhou/oss/aliyun_v4_request"}`, `{"success_action_status":"204"}`} {
		require.Contains(t, string(raw), condition)
	}
	require.NotContains(t, string(raw), "test-secret")
	require.NotContains(t, string(raw), "content-type")
	cfg.ForbidOverwrite = false
	cfg.ObjectACL = ""
	cfg.ExpiredInSec = 60
	cfg.MaxFileSize = 100
	post, err = client.SignPostPolicy("avatars/", cfg)
	require.NoError(t, err)
	require.NotContains(t, post.Fields, "x-oss-forbid-overwrite")
	require.NotContains(t, post.Fields, "x-oss-object-acl")
	raw, err = base64.StdEncoding.DecodeString(post.Fields["policy"])
	require.NoError(t, err)
	require.Contains(t, string(raw), `["content-length-range",0,100]`)
}

func TestSignPostPolicyV4_RejectsInvalidConfiguration(t *testing.T) {
	client, err := New(Config{Endpoint: "oss-cn-hangzhou.aliyuncs.com", BucketName: "private-test", Region: "cn-hangzhou", AccessKeyID: "test-ak", AccessKeySecret: "test-secret"})
	require.NoError(t, err)
	for _, cfg := range []PostPolicyConfig{{}, {ExpiredInSec: 600, MaxFileSize: 0}, {ExpiredInSec: -1, MaxFileSize: 1}, {ExpiredInSec: 1<<63 - 1, MaxFileSize: 1}, {ExpiredInSec: 600, MaxFileSize: 1, ObjectACL: "unknown"}} {
		_, err := client.SignPostPolicy("uploads/", cfg)
		require.Error(t, err)
	}
	cfg := PostPolicyConfig{ExpiredInSec: 600, MaxFileSize: 100}
	_, err = client.SignPostPolicy("", cfg)
	require.Error(t, err)
	client.region = ""
	_, err = client.SignPostPolicy("uploads/", cfg)
	require.Error(t, err)
}
