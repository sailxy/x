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
)

type ConnectPlatformConfig struct {
	ClientID     string
	ClientSecret string
}

// Docs: http://wiki.connect.qq.com/%E4%BD%BF%E7%94%A8authorization_code%E8%8E%B7%E5%8F%96access_token
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
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	OpenID       string `json:"openid"`
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
	// q.Set("need_openid", "1")

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
