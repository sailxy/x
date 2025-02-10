package oss

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	endpoint        = "oss-cn-shanghai.aliyuncs.com"
	bucketName      = ""
	accessKeyID     = ""
	accessKeySecret = ""
)

func new() (*Client, error) {
	return New(Config{
		Endpoint:        endpoint,
		BucketName:      bucketName,
		AccessKeyID:     accessKeyID,
		AccessKeySecret: accessKeySecret,
	})
}

func TestNew(t *testing.T) {
	_, err := new()
	assert.NoError(t, err)
}

func TestSignURL(t *testing.T) {
	client, err := new()
	if !assert.NoError(t, err) {
		return
	}
	resp, err := client.SignURL("1.txt", SignURLConfig{
		ContentType: "text/plain",
	})
	if assert.NoError(t, err) {
		t.Log(resp.SignedURL)
	}
}

func TestPostInfo(t *testing.T) {
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
