package oss

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func new() (*Client, error) {
	return New(Config{
		Endpoint:        os.Getenv("OSS_ENDPOINT"),
		BucketName:      os.Getenv("OSS_BUCKET"),
		AccessKeyID:     os.Getenv("OSS_ACCESS_KEY_ID"),
		AccessKeySecret: os.Getenv("OSS_ACCESS_KEY_SECRET"),
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

func TestParseHTTPMethod(t *testing.T) {
	tests := []struct {
		name    string
		input   HTTPMethod
		want    HTTPMethod
		wantErr bool
	}{
		{name: "default put", input: "", want: HTTPPut},
		{name: "put", input: HTTPPut, want: HTTPPut},
		{name: "get", input: HTTPGet, want: HTTPGet},
		{name: "invalid", input: HTTPMethod("POST"), wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseHTTPMethod(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			if assert.NoError(t, err) {
				assert.Equal(t, tt.want, got)
			}
		})
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
