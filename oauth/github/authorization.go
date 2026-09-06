package github

import (
	"errors"
	"net/url"
	"strings"

	"github.com/sailxy/x/oauth"
)

const authorizationEndpoint = "https://github.com/login/oauth/authorize"

// AuthorizationURL builds a GitHub OAuth authorization URL. State generation,
// persistence, validation, and one-time consumption remain the caller's
// responsibility.
func (c *Client) AuthorizationURL(redirectURI, state string) (string, error) {
	if strings.TrimSpace(redirectURI) == "" {
		return "", oauth.NewError(
			oauth.ProviderGitHub,
			oauth.ErrorKindInvalidInput,
			"build authorization URL",
			errors.New("redirect URI is required"),
		)
	}
	if strings.TrimSpace(state) == "" {
		return "", oauth.NewError(
			oauth.ProviderGitHub,
			oauth.ErrorKindInvalidInput,
			"build authorization URL",
			errors.New("state is required"),
		)
	}

	endpoint, err := url.Parse(authorizationEndpoint)
	if err != nil {
		return "", oauth.NewError(
			oauth.ProviderGitHub,
			oauth.ErrorKindInvalidConfig,
			"build authorization URL",
			err,
		)
	}
	query := url.Values{}
	query.Set("client_id", c.clientID)
	query.Set("redirect_uri", redirectURI)
	query.Set("state", state)
	endpoint.RawQuery = query.Encode()
	return endpoint.String(), nil
}
