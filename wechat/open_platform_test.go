package wechat

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestQRConnect(t *testing.T) {
	c := Config{AppID: "appid_123"}
	w := NewWechatOpenPlatform(c)

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
