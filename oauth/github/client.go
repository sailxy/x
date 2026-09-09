// Package github implements the GitHub OAuth web application flow.
//
// The package exchanges an authorization code and returns GitHub protocol
// data. State management, business users, token storage, and application
// sessions remain the caller's responsibility.
package github

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
	authorizationEndpoint = "https://github.com/login/oauth/authorize"
	tokenEndpoint         = "https://github.com/login/oauth/access_token"
	currentUserEndpoint   = "https://api.github.com/user"
	apiAcceptHeader       = "application/vnd.github+json"
	apiVersion            = "2022-11-28"
	userAgent             = "github.com/sailxy/x"
)

// Config configures a GitHub OAuth client.
type Config struct {
	ClientID     string
	ClientSecret oauth.SensitiveString
	HTTPClient   *http.Client
}

// Format prevents ClientSecret from leaking when the whole config is printed.
func (c Config) Format(state fmt.State, _ rune) {
	_, _ = fmt.Fprintf(state, "GitHub OAuth config for client %q", c.ClientID)
}

// Client performs GitHub OAuth protocol operations.
type Client struct {
	clientID     string
	clientSecret oauth.SensitiveString
	rest         *rest.REST
}

// Format prevents credentials and HTTP internals from leaking when the whole
// client is printed.
func (c Client) Format(state fmt.State, _ rune) {
	_, _ = fmt.Fprintf(state, "GitHub OAuth client %q", c.clientID)
}

// New creates a client and validates its credentials.
func New(config Config) (*Client, error) {
	if strings.TrimSpace(config.ClientID) == "" {
		return nil, newError(oauth.ErrorKindInvalidConfig, "create client", errors.New("client ID is required"))
	}
	if strings.TrimSpace(config.ClientSecret.Value()) == "" {
		return nil, newError(oauth.ErrorKindInvalidConfig, "create client", errors.New("client secret is required"))
	}

	httpClient := rest.NewREST()
	if config.HTTPClient != nil {
		httpClient = rest.NewRESTWithClient(config.HTTPClient)
	}
	return &Client{
		clientID:     config.ClientID,
		clientSecret: config.ClientSecret,
		rest:         httpClient,
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
		"client_id":    {c.clientID},
		"redirect_uri": {redirectURI},
		"state":        {state},
	}
	return authorizationEndpoint + "?" + query.Encode(), nil
}

// Authenticate exchanges code for a token and retrieves the authenticated
// GitHub user. The caller decides which returned fields to use or persist.
func (c *Client) Authenticate(ctx context.Context, code, redirectURI string) (*AuthenticateResult, error) {
	token, err := c.exchange(ctx, code, redirectURI)
	if err != nil {
		return nil, err
	}
	user, err := c.getCurrentUser(ctx, token.AccessToken)
	if err != nil {
		return nil, err
	}
	return &AuthenticateResult{Token: token, User: user}, nil
}

type tokenResponse struct {
	AccessToken string `json:"access_token"`
	Scope       string `json:"scope"`
	TokenType   string `json:"token_type"`
	Error       string `json:"error"`
}

func (c *Client) exchange(ctx context.Context, code, redirectURI string) (Token, error) {
	const operation = "exchange code"
	if strings.TrimSpace(code) == "" {
		return Token{}, newError(oauth.ErrorKindInvalidInput, operation, errors.New("authorization code is required"))
	}
	if strings.TrimSpace(redirectURI) == "" {
		return Token{}, newError(oauth.ErrorKindInvalidInput, operation, errors.New("redirect URI is required"))
	}

	form := url.Values{
		"client_id":     {c.clientID},
		"client_secret": {c.clientSecret.Value()},
		"code":          {code},
		"redirect_uri":  {redirectURI},
	}
	req, err := http.NewRequest(http.MethodPost, tokenEndpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return Token{}, newError(oauth.ErrorKindInvalidConfig, operation, err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", userAgent)

	response, requestErr := c.do(ctx, operation, req)
	if requestErr != nil && response == nil {
		return Token{}, requestErr
	}

	var payload tokenResponse
	if err := json.Unmarshal(response.Body, &payload); err != nil {
		if requestErr != nil {
			return Token{}, requestErr
		}
		return Token{}, newError(oauth.ErrorKindDecode, operation, err)
	}
	if payload.Error != "" {
		err := newError(oauth.ErrorKindPlatform, operation, requestErr)
		err.StatusCode = response.StatusCode
		err.Code = payload.Error
		return Token{}, err
	}
	if requestErr != nil {
		return Token{}, requestErr
	}
	if strings.TrimSpace(payload.AccessToken) == "" {
		return Token{}, newError(oauth.ErrorKindInvalidResponse, operation, errors.New("access token is missing"))
	}
	return Token{
		AccessToken: oauth.SensitiveString(payload.AccessToken),
		Scope:       payload.Scope,
		TokenType:   payload.TokenType,
	}, nil
}

func (c *Client) getCurrentUser(ctx context.Context, accessToken oauth.SensitiveString) (User, error) {
	const operation = "get current user"
	if strings.TrimSpace(accessToken.Value()) == "" {
		return User{}, newError(oauth.ErrorKindInvalidInput, operation, errors.New("access token is required"))
	}

	req, err := http.NewRequest(http.MethodGet, currentUserEndpoint, nil)
	if err != nil {
		return User{}, newError(oauth.ErrorKindInvalidConfig, operation, err)
	}
	req.Header.Set("Accept", apiAcceptHeader)
	req.Header.Set("Authorization", "Bearer "+accessToken.Value())
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("X-GitHub-Api-Version", apiVersion)

	response, err := c.do(ctx, operation, req)
	if err != nil {
		return User{}, err
	}

	var user User
	if err := json.Unmarshal(response.Body, &user); err != nil {
		return User{}, newError(oauth.ErrorKindDecode, operation, err)
	}
	if user.AccountID <= 0 {
		return User{}, newError(oauth.ErrorKindInvalidResponse, operation, errors.New("account ID is missing"))
	}
	user.Email = strings.TrimSpace(user.Email)
	return user, nil
}

func (c *Client) do(ctx context.Context, operation string, req *http.Request) (*rest.Response, error) {
	if ctx == nil {
		return nil, newError(oauth.ErrorKindInvalidInput, operation, errors.New("context is required"))
	}

	response, err := c.rest.Do(ctx, req)
	if err == nil {
		return response, nil
	}
	if errors.Is(err, rest.ErrResponseTooLarge) {
		return nil, newError(oauth.ErrorKindInvalidResponse, operation, err)
	}

	var statusErr *rest.StatusError
	if errors.As(err, &statusErr) {
		platformErr := newError(oauth.ErrorKindPlatform, operation, err)
		platformErr.StatusCode = statusErr.StatusCode
		return response, platformErr
	}
	return nil, newError(oauth.ErrorKindTransport, operation, err)
}

func newError(kind oauth.ErrorKind, operation string, cause error) *oauth.Error {
	return oauth.NewError(oauth.ProviderGitHub, kind, operation, cause)
}
