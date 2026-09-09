// Package wechat implements WeChat Open Platform OAuth login operations.
package wechat

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/sailxy/x/oauth"
	"github.com/sailxy/x/rest"
)

const (
	authorizationEndpoint = "https://open.weixin.qq.com/connect/qrconnect"
	tokenEndpoint         = "https://api.weixin.qq.com/sns/oauth2/access_token"
	userEndpoint          = "https://api.weixin.qq.com/sns/userinfo"
)

// Config configures a WeChat Open Platform OAuth client.
type Config struct {
	AppID      string
	AppSecret  oauth.SensitiveString
	HTTPClient *http.Client
}

// Format omits credentials and HTTP client internals from formatted output.
func (c Config) Format(state fmt.State, _ rune) {
	_, _ = fmt.Fprintf(state, "WeChat OAuth config for app %q", c.AppID)
}

// Client performs WeChat Open Platform OAuth protocol operations.
type Client struct {
	appID     string
	appSecret oauth.SensitiveString
	rest      *rest.REST
}

// New creates a client and validates its credentials.
func New(config Config) (*Client, error) {
	if strings.TrimSpace(config.AppID) == "" {
		return nil, newError(oauth.ErrorKindInvalidConfig, "create client", errors.New("app ID is required"))
	}
	if strings.TrimSpace(config.AppSecret.Value()) == "" {
		return nil, newError(oauth.ErrorKindInvalidConfig, "create client", errors.New("app secret is required"))
	}
	r := rest.NewREST()
	if config.HTTPClient != nil {
		r = rest.NewRESTWithClient(config.HTTPClient)
	}
	return &Client{appID: config.AppID, appSecret: config.AppSecret, rest: r}, nil
}

// AuthorizationURL builds a WeChat QR login URL. The caller manages state.
func (c *Client) AuthorizationURL(redirectURI, state string) (string, error) {
	if strings.TrimSpace(redirectURI) == "" || strings.TrimSpace(state) == "" {
		return "", newError(oauth.ErrorKindInvalidInput, "build authorization URL", errors.New("redirect URI and state are required"))
	}
	q := url.Values{"appid": {c.appID}, "redirect_uri": {redirectURI}, "response_type": {"code"}, "scope": {"snsapi_login"}, "state": {state}}
	return authorizationEndpoint + "?" + q.Encode() + "#wechat_redirect", nil
}

// Authenticate exchanges code and returns WeChat's stable OpenID and UnionID.
func (c *Client) Authenticate(ctx context.Context, code string) (*AuthenticateResult, error) {
	const operation = "exchange code"
	if strings.TrimSpace(code) == "" {
		return nil, newError(oauth.ErrorKindInvalidInput, operation, errors.New("authorization code is required"))
	}
	q := url.Values{"appid": {c.appID}, "secret": {c.appSecret.Value()}, "code": {code}, "grant_type": {"authorization_code"}}
	response, err := c.request(ctx, operation, tokenEndpoint+"?"+q.Encode())
	if err != nil {
		return nil, err
	}
	var payload tokenResponse
	if err := json.Unmarshal(response.Body, &payload); err != nil {
		return nil, newError(oauth.ErrorKindDecode, operation, err)
	}
	if payload.ErrCode != 0 {
		return nil, platformError(operation, payload.ErrCode, response.StatusCode, nil)
	}
	if strings.TrimSpace(payload.AccessToken) == "" || strings.TrimSpace(payload.OpenID) == "" || strings.TrimSpace(payload.UnionID) == "" {
		return nil, newError(oauth.ErrorKindInvalidResponse, operation, errors.New("access token, openid, or unionid is missing"))
	}
	return &AuthenticateResult{Token: Token{AccessToken: oauth.SensitiveString(payload.AccessToken), RefreshToken: oauth.SensitiveString(payload.RefreshToken), ExpiresIn: payload.ExpiresIn, Scope: payload.Scope}, Identity: Identity{OpenID: payload.OpenID, UnionID: payload.UnionID}}, nil
}

// User gets WeChat profile data only when the caller needs it.
func (c *Client) User(ctx context.Context, result *AuthenticateResult) (*User, error) {
	const operation = "get user"
	if result == nil || strings.TrimSpace(result.Token.AccessToken.Value()) == "" || strings.TrimSpace(result.Identity.OpenID) == "" {
		return nil, newError(oauth.ErrorKindInvalidInput, operation, errors.New("authentication result is incomplete"))
	}
	q := url.Values{"access_token": {result.Token.AccessToken.Value()}, "openid": {result.Identity.OpenID}}
	response, err := c.request(ctx, operation, userEndpoint+"?"+q.Encode())
	if err != nil {
		return nil, err
	}
	var user User
	if err := json.Unmarshal(response.Body, &user); err != nil {
		return nil, newError(oauth.ErrorKindDecode, operation, err)
	}
	if user.ErrCode != 0 {
		return nil, platformError(operation, user.ErrCode, response.StatusCode, nil)
	}
	return &user, nil
}

type tokenResponse struct {
	AccessToken  string `json:"access_token"`
	ExpiresIn    int64  `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	OpenID       string `json:"openid"`
	Scope        string `json:"scope"`
	UnionID      string `json:"unionid"`
	ErrCode      int    `json:"errcode"`
}

func (c *Client) request(ctx context.Context, operation, address string) (*rest.Response, error) {
	if ctx == nil {
		return nil, newError(oauth.ErrorKindInvalidInput, operation, errors.New("context is required"))
	}
	response, err := c.rest.Get(ctx, address)
	if err == nil {
		return response, nil
	}
	if errors.Is(err, rest.ErrResponseTooLarge) {
		return nil, newError(oauth.ErrorKindInvalidResponse, operation, err)
	}
	var statusErr *rest.StatusError
	if errors.As(err, &statusErr) {
		return response, platformError(operation, 0, statusErr.StatusCode, err)
	}
	return nil, newError(oauth.ErrorKindTransport, operation, err)
}

func platformError(operation string, code, status int, cause error) *oauth.Error {
	err := newError(oauth.ErrorKindPlatform, operation, cause)
	err.Code = fmt.Sprint(code)
	err.StatusCode = status
	return err
}
func newError(kind oauth.ErrorKind, operation string, cause error) *oauth.Error {
	return oauth.NewError(oauth.ProviderWechat, kind, operation, cause)
}
