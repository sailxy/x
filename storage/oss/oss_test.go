package oss

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	requireAliyunOSSEnv(t)

	_, err := new()
	assert.NoError(t, err)
}

func TestParseObjectKeyFromURL(t *testing.T) {
	tests := []struct {
		name    string
		rawURL  string
		want    string
		wantErr bool
	}{
		{
			name:   "nested path",
			rawURL: "https://bucket.oss-cn-hangzhou.aliyuncs.com/path/to/file.txt",
			want:   "path/to/file.txt",
		},
		{
			name:   "ignores query and fragment",
			rawURL: "https://bucket.oss-cn-hangzhou.aliyuncs.com/path/to/file.txt?Expires=1&OSSAccessKeyId=id#section",
			want:   "path/to/file.txt",
		},
		{
			name:   "decodes escaped characters",
			rawURL: "https://bucket.oss-cn-hangzhou.aliyuncs.com/path/%E4%B8%AD%E6%96%87%20file.txt",
			want:   "path/中文 file.txt",
		},
		{
			name:    "malformed url",
			rawURL:  "http://%",
			wantErr: true,
		},
		{
			name:    "relative path is not url",
			rawURL:  "path/to/file.txt",
			wantErr: true,
		},
		{
			name:    "empty path",
			rawURL:  "https://bucket.oss-cn-hangzhou.aliyuncs.com",
			wantErr: true,
		},
		{
			name:    "root path",
			rawURL:  "https://bucket.oss-cn-hangzhou.aliyuncs.com/",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseObjectKeyFromURL(tt.rawURL)
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
