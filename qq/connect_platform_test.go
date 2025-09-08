package qq

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func newConnectPlatform() *ConnectPlatform {
	return NewConnectPlatform(ConnectPlatformConfig{
		ClientID:     os.Getenv("QQ_WEB_CLIENT_ID"),
		ClientSecret: os.Getenv("QQ_WEB_CLIENT_SECRET"),
	})
}

func TestGetAuthURL(t *testing.T) {
	c := newConnectPlatform()
	got, err := c.GetAuthURL("https://example.com/callback", "xyz")
	if assert.NoError(t, err) {
		t.Log(got)
	}
}

func TestGetAccessToken(t *testing.T) {
	c := newConnectPlatform()
	got, err := c.GetAccessToken("code_123", "https://example.com/callback")
	if assert.NoError(t, err) {
		t.Log(got)
	}
}

func TestGetUnionID(t *testing.T) {
	c := newConnectPlatform()
	got, err := c.GetUnionID("token_123")
	if assert.NoError(t, err) {
		t.Log(got)
	}
}

func TestGetUserInfo(t *testing.T) {
	c := newConnectPlatform()
	got, err := c.GetUserInfo("token_123", "openid_123")
	if assert.NoError(t, err) {
		t.Log(got)
	}
}
