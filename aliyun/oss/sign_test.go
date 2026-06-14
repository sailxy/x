package oss

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

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
