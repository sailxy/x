package github

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/sailxy/x/oauth"
)

const (
	currentUserEndpoint = "https://api.github.com/user"
	apiAcceptHeader     = "application/vnd.github+json"
	apiVersion          = "2022-11-28"
)

func (c *Client) getCurrentUser(ctx context.Context, accessToken oauth.SensitiveString) (User, error) {
	const operation = "get current user"
	if strings.TrimSpace(accessToken.Value()) == "" {
		return User{}, oauth.NewError(
			oauth.ProviderGitHub,
			oauth.ErrorKindInvalidInput,
			operation,
			errors.New("access token is required"),
		)
	}
	req, err := http.NewRequest(http.MethodGet, currentUserEndpoint, nil)
	if err != nil {
		return User{}, oauth.NewError(
			oauth.ProviderGitHub,
			oauth.ErrorKindInvalidConfig,
			operation,
			err,
		)
	}
	req.Header.Set("Accept", apiAcceptHeader)
	req.Header.Set("Authorization", "Bearer "+accessToken.Value())
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("X-GitHub-Api-Version", apiVersion)

	resp, err := c.do(ctx, operation, req)
	if err != nil {
		return User{}, err
	}

	var payload User
	if err := json.Unmarshal(resp.Body, &payload); err != nil {
		return User{}, oauth.NewError(
			oauth.ProviderGitHub,
			oauth.ErrorKindDecode,
			operation,
			err,
		)
	}
	if payload.AccountID <= 0 {
		return User{}, oauth.NewError(
			oauth.ProviderGitHub,
			oauth.ErrorKindInvalidResponse,
			operation,
			errors.New("account ID is missing"),
		)
	}
	payload.Email = strings.TrimSpace(payload.Email)
	return payload, nil
}
