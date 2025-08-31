package wechat

import (
	"net/url"
)

type Config struct {
	AppID string
}

type WechatOpenPlatform struct {
	appID string
}

const openPlatformQRConnectURL = "https://open.weixin.qq.com/connect/qrconnect"

// NewWechatOpenPlatform create a new wechat open platform instance.
func NewWechatOpenPlatform(c *Config) *WechatOpenPlatform {
	return &WechatOpenPlatform{
		appID: c.AppID,
	}
}

// QRConnect generate a QR connect URL for the wechat open platform.
func (w *WechatOpenPlatform) QRConnect(redirectURI string, state string) (string, error) {
	u, err := url.Parse(openPlatformQRConnectURL)
	if err != nil {
		return "", err
	}

	q := url.Values{}
	q.Set("appid", w.appID)
	q.Set("redirect_uri", redirectURI)
	q.Set("response_type", "code")
	q.Set("scope", "snsapi_login")
	q.Set("state", state)

	u.RawQuery = q.Encode()
	return u.String(), nil
}
