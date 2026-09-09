// Package qq implements QQ OAuth login protocol operations.
package qq

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/sailxy/x/oauth"
	"github.com/sailxy/x/rest"
)

const (
	authorizationEndpoint = "https://graph.qq.com/oauth2.0/authorize"
	tokenEndpoint         = "https://graph.qq.com/oauth2.0/token"
	identityEndpoint      = "https://graph.qq.com/oauth2.0/me"
	userEndpoint          = "https://graph.qq.com/user/get_user_info"
)

// Config configures a QQ OAuth client.
type Config struct {
	ClientID     string
	ClientSecret oauth.SensitiveString
	HTTPClient   *http.Client
}

// Format omits credentials and HTTP client internals from formatted output.
func (c Config) Format(state fmt.State, _ rune) {
	_, _ = fmt.Fprintf(state, "QQ OAuth config for client %q", c.ClientID)
}

// Client performs QQ OAuth protocol operations.
type Client struct {
	clientID     string
	clientSecret oauth.SensitiveString
	rest         *rest.REST
}

// Format omits credentials and HTTP client internals from formatted output.
func (c Client) Format(state fmt.State, _ rune) {
	_, _ = fmt.Fprintf(state, "QQ OAuth client %q", c.clientID)
}

// New creates a client and validates its credentials.
func New(config Config) (*Client, error) {
	if strings.TrimSpace(config.ClientID) == "" {
		return nil, newError(oauth.ErrorKindInvalidConfig, "create client", errors.New("client ID is required"))
	}
	if strings.TrimSpace(config.ClientSecret.Value()) == "" {
		return nil, newError(oauth.ErrorKindInvalidConfig, "create client", errors.New("client secret is required"))
	}

	restClient := rest.NewREST()
	if config.HTTPClient != nil {
		restClient = rest.NewRESTWithClient(config.HTTPClient)
	}
	return &Client{
		clientID:     config.ClientID,
		clientSecret: config.ClientSecret,
		rest:         restClient,
	}, nil
}

// AuthorizationURL builds the URL to which the user's browser is redirected.
// The caller creates, stores, and consumes state.
func (c *Client) AuthorizationURL(redirectURI, state string) (string, error) {
	if strings.TrimSpace(redirectURI) == "" {
		return "", newError(oauth.ErrorKindInvalidInput, "build authorization URL", errors.New("redirect URI is required"))
	}
	if strings.TrimSpace(state) == "" {
		return "", newError(oauth.ErrorKindInvalidInput, "build authorization URL", errors.New("state is required"))
	}

	query := url.Values{
		"response_type": {"code"},
		"client_id":     {c.clientID},
		"redirect_uri":  {redirectURI},
		"state":         {state},
	}
	return authorizationEndpoint + "?" + query.Encode(), nil
}

// Authenticate exchanges code and returns QQ's stable OpenID and UnionID.
func (c *Client) Authenticate(ctx context.Context, code, redirectURI string) (*AuthenticateResult, error) {
	token, err := c.exchange(ctx, code, redirectURI)
	if err != nil {
		return nil, err
	}
	identity, err := c.identity(ctx, token.AccessToken)
	if err != nil {
		return nil, err
	}
	return &AuthenticateResult{Token: token, Identity: identity}, nil
}

// User gets QQ profile data only when the caller needs it.
func (c *Client) User(ctx context.Context, result *AuthenticateResult) (*User, error) {
	const operation = "get user"
	if result == nil || strings.TrimSpace(result.Token.AccessToken.Value()) == "" || strings.TrimSpace(result.Identity.OpenID) == "" {
		return nil, newError(oauth.ErrorKindInvalidInput, operation, errors.New("authentication result is incomplete"))
	}
	query := url.Values{
		"access_token":       {result.Token.AccessToken.Value()},
		"oauth_consumer_key": {c.clientID},
		"openid":             {result.Identity.OpenID},
	}
	response, err := c.request(ctx, operation, userEndpoint+"?"+query.Encode())
	if err != nil {
		return nil, err
	}
	var user User
	if err := json.Unmarshal(response.Body, &user); err != nil {
		return nil, newError(oauth.ErrorKindDecode, operation, err)
	}
	if user.Ret != 0 {
		return nil, platformError(operation, strconv.Itoa(user.Ret), response.StatusCode, nil)
	}
	return &user, nil
}

type tokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    string `json:"expires_in"`
	Error        int    `json:"error"`
}

type identityResponse struct {
	OpenID  string `json:"openid"`
	UnionID string `json:"unionid"`
	Error   int    `json:"error"`
}

func (c *Client) exchange(ctx context.Context, code, redirectURI string) (Token, error) {
	const operation = "exchange code"
	if strings.TrimSpace(code) == "" || strings.TrimSpace(redirectURI) == "" {
		return Token{}, newError(oauth.ErrorKindInvalidInput, operation, errors.New("authorization code and redirect URI are required"))
	}
	query := url.Values{
		"grant_type":    {"authorization_code"},
		"client_id":     {c.clientID},
		"client_secret": {c.clientSecret.Value()},
		"code":          {code},
		"redirect_uri":  {redirectURI},
		"fmt":           {"json"},
		"need_openid":   {"1"},
	}
	response, err := c.request(ctx, operation, tokenEndpoint+"?"+query.Encode())
	if err != nil {
		return Token{}, err
	}
	var payload tokenResponse
	if err := json.Unmarshal(response.Body, &payload); err != nil {
		return Token{}, newError(oauth.ErrorKindDecode, operation, err)
	}
	if payload.Error != 0 {
		return Token{}, platformError(operation, strconv.Itoa(payload.Error), response.StatusCode, nil)
	}
	if strings.TrimSpace(payload.AccessToken) == "" {
		return Token{}, newError(oauth.ErrorKindInvalidResponse, operation, errors.New("access token is missing"))
	}
	return Token{AccessToken: oauth.SensitiveString(payload.AccessToken), RefreshToken: oauth.SensitiveString(payload.RefreshToken), ExpiresIn: payload.ExpiresIn}, nil
}

func (c *Client) identity(ctx context.Context, token oauth.SensitiveString) (Identity, error) {
	const operation = "get identity"
	query := url.Values{"access_token": {token.Value()}, "unionid": {"1"}, "fmt": {"json"}}
	response, err := c.request(ctx, operation, identityEndpoint+"?"+query.Encode())
	if err != nil {
		return Identity{}, err
	}
	var payload identityResponse
	if err := json.Unmarshal(response.Body, &payload); err != nil {
		return Identity{}, newError(oauth.ErrorKindDecode, operation, err)
	}
	if payload.Error != 0 {
		return Identity{}, platformError(operation, strconv.Itoa(payload.Error), response.StatusCode, nil)
	}
	if strings.TrimSpace(payload.OpenID) == "" || strings.TrimSpace(payload.UnionID) == "" {
		return Identity{}, newError(oauth.ErrorKindInvalidResponse, operation, errors.New("openid or unionid is missing"))
	}
	return Identity{OpenID: payload.OpenID, UnionID: payload.UnionID}, nil
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
		return response, platformError(operation, "", statusErr.StatusCode, err)
	}
	return nil, newError(oauth.ErrorKindTransport, operation, err)
}

func platformError(operation, code string, status int, cause error) *oauth.Error {
	err := newError(oauth.ErrorKindPlatform, operation, cause)
	err.Code = code
	err.StatusCode = status
	return err
}

func newError(kind oauth.ErrorKind, operation string, cause error) *oauth.Error {
	return oauth.NewError(oauth.ProviderQQ, kind, operation, cause)
}
