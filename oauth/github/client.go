// Package github implements the GitHub OAuth web application flow.
//
// It exchanges provider credentials and returns GitHub protocol data only. It
// does not generate or validate application state, match business users, store
// tokens, or issue application sessions.
package github

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/sailxy/x/oauth"
	"github.com/sailxy/x/rest"
)

const userAgent = "github.com/sailxy/x"

// Config configures a GitHub OAuth client.
type Config struct {
	ClientID     string
	ClientSecret oauth.SensitiveString
	HTTPClient   *http.Client
}

// Format omits the Client Secret and HTTP client internals from formatted
// configuration output.
func (c Config) Format(state fmt.State, _ rune) {
	_, _ = fmt.Fprintf(state, "GitHub OAuth config for client %q", c.ClientID)
}

// Client performs GitHub OAuth protocol operations.
type Client struct {
	clientID     string
	clientSecret oauth.SensitiveString
	rest         *rest.REST
}

// New creates a GitHub OAuth client and validates required credentials.
func New(config Config) (*Client, error) {
	if strings.TrimSpace(config.ClientID) == "" {
		return nil, oauth.NewError(
			oauth.ProviderGitHub,
			oauth.ErrorKindInvalidConfig,
			"create client",
			errors.New("client ID is required"),
		)
	}
	if strings.TrimSpace(config.ClientSecret.Value()) == "" {
		return nil, oauth.NewError(
			oauth.ProviderGitHub,
			oauth.ErrorKindInvalidConfig,
			"create client",
			errors.New("client secret is required"),
		)
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

// Authenticate exchanges code for a token and uses that token to retrieve the
// authenticated GitHub user. It returns all parsed token and user data without
// applying business account rules.
func (c *Client) Authenticate(ctx context.Context, code, redirectURI string) (*AuthenticateResult, error) {
	token, err := c.exchange(ctx, code, redirectURI)
	if err != nil {
		return nil, err
	}
	user, err := c.getCurrentUser(ctx, token.AccessToken)
	if err != nil {
		return nil, err
	}
	return &AuthenticateResult{
		Token: token,
		User:  user,
	}, nil
}

func (c *Client) do(ctx context.Context, operation string, req *http.Request) (*rest.Response, error) {
	if ctx == nil {
		return nil, oauth.NewError(
			oauth.ProviderGitHub,
			oauth.ErrorKindInvalidInput,
			operation,
			errors.New("context is required"),
		)
	}

	resp, err := c.rest.Do(ctx, req)
	if err == nil {
		return resp, nil
	}
	if errors.Is(err, rest.ErrResponseTooLarge) {
		return nil, oauth.NewError(
			oauth.ProviderGitHub,
			oauth.ErrorKindInvalidResponse,
			operation,
			err,
		)
	}

	var statusErr *rest.StatusError
	if errors.As(err, &statusErr) {
		platformErr := oauth.NewError(
			oauth.ProviderGitHub,
			oauth.ErrorKindPlatform,
			operation,
			err,
		)
		platformErr.StatusCode = statusErr.StatusCode
		return resp, platformErr
	}
	return nil, oauth.NewError(
		oauth.ProviderGitHub,
		oauth.ErrorKindTransport,
		operation,
		err,
	)
}

// Format omits credentials and HTTP client internals from formatted output.
func (c Client) Format(state fmt.State, _ rune) {
	_, _ = fmt.Fprintf(state, "GitHub OAuth client %q", c.clientID)
}
