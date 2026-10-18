package oss

import (
	"net/http"
	"net/url"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSignURL_ResponseContentDisposition(t *testing.T) {
	client, err := New(Config{
		Endpoint: "https://oss-cn-hangzhou.aliyuncs.com", BucketName: "example-bucket",
		AccessKeyID: "test-access-key", AccessKeySecret: "test-secret",
	})
	require.NoError(t, err)
	for _, tt := range []struct {
		name string
		cfg  SignURLConfig
	}{
		{name: "default upload", cfg: SignURLConfig{}},
		{name: "ordinary download", cfg: SignURLConfig{HTTPMethod: HTTPGet}},
		{name: "named download", cfg: SignURLConfig{
			HTTPMethod: HTTPGet, ExpiredInSec: 60,
			ResponseContentDisposition: `attachment; filename="example-export.zip"`,
		}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			before := time.Now().Unix()
			resp, err := client.SignURL("exports/attempt-1.zip", tt.cfg)
			require.NoError(t, err)
			parsed, err := url.Parse(resp.SignedURL)
			require.NoError(t, err)
			query := parsed.Query()
			assert.Equal(t, tt.cfg.ResponseContentDisposition, query.Get("response-content-disposition"))
			assert.Equal(t, tt.cfg.ResponseContentDisposition != "", query.Has("response-content-disposition"))
			assert.Equal(t, "/exports/attempt-1.zip", parsed.Path)
			assert.NotEmpty(t, query.Get("Signature"))
			wantTTL := tt.cfg.ExpiredInSec
			if wantTTL == 0 {
				wantTTL = defaultExpiredInSec
			}
			assert.Equal(t, wantTTL, resp.ExpiredInSec)
			expires, err := strconv.ParseInt(query.Get("Expires"), 10, 64)
			require.NoError(t, err)
			assert.GreaterOrEqual(t, expires, before+wantTTL)
			assert.LessOrEqual(t, expires, time.Now().Unix()+wantTTL)
		})
	}
}

func TestSignURL(t *testing.T) {
	requireAliyunOSSEnv(t)

	client, err := new()
	if !assert.NoError(t, err) {
		return
	}
	resp, err := client.SignURL("test.txt", SignURLConfig{
		HTTPMethod: HTTPGet,
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

func TestSignURL_V4PreservesRequiredHeadersAndOptions(t *testing.T) {
	c, err := New(Config{Endpoint: "https://oss-cn-hangzhou.aliyuncs.com", BucketName: "private-test", Region: "cn-hangzhou", AccessKeyID: "test-ak", AccessKeySecret: "test-secret"})
	require.NoError(t, err)
	signed, err := c.SignURL("proof.pdf", SignURLConfig{HTTPMethod: HTTPPut, ExpiredInSec: 60, ContentType: "application/pdf", Callback: "callback", CallbackVar: "vars"})
	require.NoError(t, err)
	parsed, err := url.Parse(signed.SignedURL)
	require.NoError(t, err)
	require.NotEmpty(t, parsed.Query().Get("x-oss-signature"))
	require.Equal(t, "60", parsed.Query().Get("x-oss-expires"))
	headers := http.Header{}
	for name, value := range signed.SignedHeaders {
		headers.Set(name, value)
	}
	require.Equal(t, "application/pdf", headers.Get("Content-Type"))
	require.Equal(t, "callback", headers.Get("x-oss-callback"))
	require.Equal(t, "vars", headers.Get("x-oss-callback-var"))
	_, err = c.SignURL("proof.pdf", SignURLConfig{ExpiredInSec: -1})
	require.Error(t, err)
	_, err = c.SignURL("proof.pdf", SignURLConfig{ExpiredInSec: 1<<63 - 1})
	require.Error(t, err)
}
