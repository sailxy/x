package wechat

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/sailxy/x/rest"
)

const (
	openPlatformQRConnectURL      = "https://open.weixin.qq.com/connect/qrconnect"
	openPlatformGetAccessTokenURL = "https://api.weixin.qq.com/sns/oauth2/access_token"
	openPlatformGetUserInfoURL    = "https://api.weixin.qq.com/sns/userinfo"
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

type getUserInfoResp struct {
	OpenID     string `json:"openid"`
	Nickname   string `json:"nickname"`
	Sex        int    `json:"sex"`
	Province   string `json:"province"`
	City       string `json:"city"`
	Country    string `json:"country"`
	HeadImgURL string `json:"headimgurl"`
	UnionID    string `json:"unionid"`
}

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

// GetUserInfo get user info through access_token and openid.
func (w *WechatOpenPlatform) GetUserInfo(accessToken, openID string) (*getUserInfoResp, error) {
	u, err := url.Parse(openPlatformGetUserInfoURL)
	if err != nil {
		return nil, err
	}

	q := url.Values{}
	q.Set("access_token", accessToken)
	q.Set("openid", openID)

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
	var respData getUserInfoResp
	err = json.Unmarshal(resp.Bytes(), &respData)
	if err != nil {
		return nil, err
	}
	return &respData, nil
}
