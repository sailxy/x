package oss

import (
	"os"
	"strings"
	"testing"
)

func new() (*Client, error) {
	return New(Config{
		Endpoint:        os.Getenv("ALIYUN_OSS_ENDPOINT"),
		BucketName:      os.Getenv("ALIYUN_OSS_BUCKET"),
		AccessKeyID:     os.Getenv("ALIYUN_OSS_ACCESS_KEY_ID"),
		AccessKeySecret: os.Getenv("ALIYUN_OSS_ACCESS_KEY_SECRET"),
	})
}

func requireAliyunOSSEnv(t *testing.T) {
	t.Helper()

	required := []string{
		"ALIYUN_OSS_ENDPOINT",
		"ALIYUN_OSS_BUCKET",
		"ALIYUN_OSS_ACCESS_KEY_ID",
		"ALIYUN_OSS_ACCESS_KEY_SECRET",
	}
	for _, key := range required {
		if strings.TrimSpace(os.Getenv(key)) == "" {
			t.Skipf("%s is not set", key)
		}
	}
}
