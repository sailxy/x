package github

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strings"

	"github.com/sailxy/x/oauth"
)

const tokenEndpoint = "https://github.com/login/oauth/access_token"

type tokenResponse struct {
	AccessToken string `json:"access_token"`
	Scope       string `json:"scope"`
	TokenType   string `json:"token_type"`
	Error       string `json:"error"`
}

func (c *Client) exchange(ctx context.Context, code, redirectURI string) (Token, error) {
	if strings.TrimSpace(code) == "" {
		return Token{}, oauth.NewError(
			oauth.ProviderGitHub,
			oauth.ErrorKindInvalidInput,
			"exchange code",
			errors.New("authorization code is required"),
		)
	}
	if strings.TrimSpace(redirectURI) == "" {
		return Token{}, oauth.NewError(
			oauth.ProviderGitHub,
			oauth.ErrorKindInvalidInput,
			"exchange code",
			errors.New("redirect URI is required"),
		)
	}

	form := url.Values{}
	form.Set("client_id", c.clientID)
	form.Set("client_secret", c.clientSecret.Value())
	form.Set("code", code)
	form.Set("redirect_uri", redirectURI)
	req, err := http.NewRequest(http.MethodPost, tokenEndpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return Token{}, oauth.NewError(
			oauth.ProviderGitHub,
			oauth.ErrorKindInvalidConfig,
			"exchange code",
			err,
		)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", userAgent)

	resp, requestErr := c.do(ctx, "exchange code", req)
	if requestErr != nil && resp == nil {
		return Token{}, requestErr
	}

	var payload tokenResponse
	if err := json.Unmarshal(resp.Body, &payload); err != nil {
		if requestErr != nil {
			return Token{}, requestErr
		}
		return Token{}, oauth.NewError(
			oauth.ProviderGitHub,
			oauth.ErrorKindDecode,
			"exchange code",
			err,
		)
	}
	if payload.Error != "" {
		platformErr := oauth.NewError(
			oauth.ProviderGitHub,
			oauth.ErrorKindPlatform,
			"exchange code",
			requestErr,
		)
		platformErr.StatusCode = resp.StatusCode
		platformErr.Code = payload.Error
		return Token{}, platformErr
	}
	if requestErr != nil {
		return Token{}, requestErr
	}
	if strings.TrimSpace(payload.AccessToken) == "" {
		return Token{}, oauth.NewError(
			oauth.ProviderGitHub,
			oauth.ErrorKindInvalidResponse,
			"exchange code",
			errors.New("access token is missing"),
		)
	}
	return Token{
		AccessToken: oauth.SensitiveString(payload.AccessToken),
		Scope:       payload.Scope,
		TokenType:   payload.TokenType,
	}, nil
}
