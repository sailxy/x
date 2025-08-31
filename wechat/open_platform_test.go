package wechat

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func newWechatOpenPlatform() *WechatOpenPlatform {
	return NewWechatOpenPlatform(Config{
		AppID:     os.Getenv("WECHAT_OPEN_WEB_APP_ID"),
		AppSecret: os.Getenv("WECHAT_OPEN_WEB_APP_SECRET"),
	})
}

func TestQRConnect(t *testing.T) {
	w := newWechatOpenPlatform()

	tests := []struct {
		name        string
		redirectURI string
		state       string
	}{
		{"basic", "https://example.com/callback", "xyz"},
		{"with_query_and_space", "https://example.com/callback?from=weixin&x=1 2", "state-123"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := w.QRConnect(tt.redirectURI, tt.state)
			if assert.NoError(t, err) {
				t.Log(got)
			}
		})
	}
}

func TestGetAccessToken(t *testing.T) {
	w := newWechatOpenPlatform()

	tests := []struct {
		name    string
		code    string
		wantErr bool
	}{
		{"basic", "code123", false},
		{"invalid_code", "invalid_code", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := w.GetAccessToken(tt.code)
			if tt.wantErr {
				assert.Error(t, err)
			} else if assert.NoError(t, err) {
				t.Log(got)
			}
		})
	}
}
