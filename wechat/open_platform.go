package wechat

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/sailxy/x/rest"
)

type Config struct {
	AppID     string
	AppSecret string
}

type WechatOpenPlatform struct {
	appID     string
	appSecret string

	client *rest.REST
}

type errResp struct {
	ErrCode int    `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
	RID     string `json:"rid"`
}

type getAccessTokenResp struct {
	AccessToken  string `json:"access_token"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	OpenID       string `json:"openid"`
	Scope        string `json:"scope"`
	UnionID      string `json:"unionid"`
}

const openPlatformQRConnectURL = "https://open.weixin.qq.com/connect/qrconnect"
const openPlatformGetAccessTokenURL = "https://api.weixin.qq.com/sns/oauth2/access_token"

// NewWechatOpenPlatform create a new wechat open platform instance.
func NewWechatOpenPlatform(c Config) *WechatOpenPlatform {
	return &WechatOpenPlatform{
		appID:     c.AppID,
		appSecret: c.AppSecret,
		client:    rest.NewREST(),
	}
}

// QRConnect generate a QR connect URL for the wechat open platform.
// After the user allows authorization, they will be redirected to the redirect_uri URL with code and state parameters.
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

// GetAccessToken get access_token through code.
func (w *WechatOpenPlatform) GetAccessToken(code string) (*getAccessTokenResp, error) {
	u, err := url.Parse(openPlatformGetAccessTokenURL)
	if err != nil {
		return nil, err
	}

	q := url.Values{}
	q.Set("appid", w.appID)
	q.Set("secret", w.appSecret)
	q.Set("code", code)
	q.Set("grant_type", "authorization_code")

	u.RawQuery = q.Encode()
	resp, err := w.client.Get(u.String())
	if err != nil {
		return nil, err
	}

	// Check error first.
	var errResp errResp
	err = json.Unmarshal(resp.Bytes(), &errResp)
	if err != nil {
		return nil, err
	}
	if errResp.ErrCode != 0 {
		return nil, fmt.Errorf("errcode: %d, errmsg: %s, rid: %s", errResp.ErrCode, errResp.ErrMsg, errResp.RID)
	}

	// Parse response.
	var respData getAccessTokenResp
	err = json.Unmarshal(resp.Bytes(), &respData)
	if err != nil {
		return nil, err
	}
	return &respData, nil
}
