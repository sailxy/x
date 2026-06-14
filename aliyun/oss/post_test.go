package oss

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPostInfo(t *testing.T) {
	requireAliyunOSSEnv(t)

	client, err := new()
	if !assert.NoError(t, err) {
		return
	}
	resp, err := client.PostInfo("test")
	if assert.NoError(t, err) {
		t.Log("access key id", resp.AccessKeyId)
		t.Log("callback", resp.Callback)
		t.Log("directory", resp.Directory)
		t.Log("expire", resp.Expire)
		t.Log("host", resp.Host)
		t.Log("policy", resp.Policy)
		t.Log("signature", resp.Signature)
	}
}

func TestSignPostPolicy(t *testing.T) {
	policy := base64.StdEncoding.EncodeToString([]byte(`{"expiration":"2026-04-23T00:00:00Z"}`))
	secret := "secret"

	mac := hmac.New(sha1.New, []byte(secret))
	_, err := mac.Write([]byte(policy))
	assert.NoError(t, err)

	want := base64.StdEncoding.EncodeToString(mac.Sum(nil))
	assert.Equal(t, want, signPostPolicy(policy, secret))
}
