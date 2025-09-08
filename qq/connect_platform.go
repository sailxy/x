package qq

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/sailxy/x/rest"
)

const (
	connectPlatformAuthURL           = "https://graph.qq.com/oauth2.0/authorize"
	connectPlatformGetAccessTokenURL = "https://graph.qq.com/oauth2.0/token"
	connectPlatformGetUnionIDURL     = "https://graph.qq.com/oauth2.0/me"
	connectPlatformGetUserInfoURL    = "https://graph.qq.com/user/get_user_info"
)

type ConnectPlatformConfig struct {
	ClientID     string
	ClientSecret string
}

// Docs:
// https://wiki.connect.qq.com/%E4%BD%BF%E7%94%A8authorization_code%E8%8E%B7%E5%8F%96access_token
// https://wiki.connect.qq.com/%E5%85%AC%E5%85%B1%E8%BF%94%E5%9B%9E%E7%A0%81%E8%AF%B4%E6%98%8E
type ConnectPlatform struct {
	clientID     string
	clientSecret string

	client *rest.REST
}

type errResp struct {
	Error            int    `json:"error"`
	ErrorDescription string `json:"error_description"`
}

type getAccessTokenResp struct {
	AccessToken  string `json:"access_token"`
	ExpiresIn    string `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	OpenID       string `json:"openid"`
}

type getUnionIDResp struct {
	ClientID string `json:"client_id"`
	OpenID   string `json:"openid"`
	UnionID  string `json:"unionid"`
}

type getUserInfoResp struct {
	Ret          int    `json:"ret"`            // Return code.
	Msg          string `json:"msg"`            // If ret < 0, there will be an error message prompt, and the returned data is encoded in UTF-8.
	IsLost       int    `json:"is_lost"`        // Whether there is data loss.
	Nickname     string `json:"nickname"`       // The nickname of the user in the QQ space.
	FigureURL    string `json:"figureurl"`      // The URL of the QQ space avatar of 30×30 pixels.
	FigureURL1   string `json:"figureurl_1"`    // The URL of the QQ space avatar of 50×50 pixels.
	FigureURL2   string `json:"figureurl_2"`    // The URL of the QQ space avatar of 100×100 pixels.
	FigureURLQQ1 string `json:"figureurl_qq_1"` // The URL of the QQ avatar of 40×40 pixels.
	FigureURLQQ2 string `json:"figureurl_qq_2"` // The URL of the QQ avatar of 100×100 pixels.
}

func NewConnectPlatform(c ConnectPlatformConfig) *ConnectPlatform {
	return &ConnectPlatform{
		clientID:     c.ClientID,
		clientSecret: c.ClientSecret,
		client:       rest.NewREST(),
	}
}

// GetAuthURL get auth URL.
// After the user allows authorization, they will be redirected to the redirect_uri URL with code and state parameters.
func (c *ConnectPlatform) GetAuthURL(redirectURI string, state string) (string, error) {
	u, err := url.Parse(connectPlatformAuthURL)
	if err != nil {
		return "", err
	}

	q := url.Values{}
	q.Set("response_type", "code")
	q.Set("client_id", c.clientID)
	q.Set("redirect_uri", redirectURI)
	q.Set("state", state)

	u.RawQuery = q.Encode()
	return u.String(), nil
}

// GetAccessToken get access token through code.
func (c *ConnectPlatform) GetAccessToken(code string, redirectURI string) (*getAccessTokenResp, error) {
	u, err := url.Parse(connectPlatformGetAccessTokenURL)
	if err != nil {
		return nil, err
	}

	q := url.Values{}
	q.Set("grant_type", "authorization_code")
	q.Set("client_id", c.clientID)
	q.Set("client_secret", c.clientSecret)
	q.Set("code", code)
	q.Set("redirect_uri", redirectURI)
	q.Set("fmt", "json")
	q.Set("need_openid", "1")

	u.RawQuery = q.Encode()
	resp, err := c.client.Get(u.String())
	if err != nil {
		return nil, err
	}

	// Check error first.
	var errResp errResp
	err = json.Unmarshal(resp.Bytes(), &errResp)
	if err != nil {
		return nil, err
	}
	if errResp.Error != 0 {
		return nil, fmt.Errorf("error: %d, error_description: %s", errResp.Error, errResp.ErrorDescription)
	}

	// Parse response.
	var respData getAccessTokenResp
	err = json.Unmarshal(resp.Bytes(), &respData)
	if err != nil {
		return nil, err
	}

	return &respData, nil
}

// GetUnionID get unionid through access_token.
func (c *ConnectPlatform) GetUnionID(accessToken string) (*getUnionIDResp, error) {
	u, err := url.Parse(connectPlatformGetUnionIDURL)
	if err != nil {
		return nil, err
	}

	q := url.Values{}
	q.Set("access_token", accessToken)
	q.Set("unionid", "1")
	q.Set("fmt", "json")

	u.RawQuery = q.Encode()
	resp, err := c.client.Get(u.String())
	if err != nil {
		return nil, err
	}

	// Check error first.
	var errResp errResp
	err = json.Unmarshal(resp.Bytes(), &errResp)
	if err != nil {
		return nil, err
	}
	if errResp.Error != 0 {
		return nil, fmt.Errorf("error: %d, error_description: %s", errResp.Error, errResp.ErrorDescription)
	}

	// Parse response.
	var respData getUnionIDResp
	err = json.Unmarshal(resp.Bytes(), &respData)
	if err != nil {
		return nil, err
	}
	return &respData, nil
}

// GetUserInfo get user info through access_token and openid.
func (c *ConnectPlatform) GetUserInfo(accessToken, openID string) (*getUserInfoResp, error) {
	u, err := url.Parse(connectPlatformGetUserInfoURL)
	if err != nil {
		return nil, err
	}

	q := url.Values{}
	q.Set("access_token", accessToken)
	q.Set("oauth_consumer_key", c.clientID)
	q.Set("openid", openID)

	u.RawQuery = q.Encode()
	resp, err := c.client.Get(u.String())
	if err != nil {
		return nil, err
	}

	// Parse response.
	var respData getUserInfoResp
	err = json.Unmarshal(resp.Bytes(), &respData)
	if err != nil {
		return nil, err
	}

	if respData.Ret < 0 {
		return nil, fmt.Errorf("ret: %d, msg: %s", respData.Ret, respData.Msg)
	}

	return &respData, nil
}
