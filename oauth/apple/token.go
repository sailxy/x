package apple

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strings"

	"github.com/sailxy/x/oauth"
)

const (
	tokenEndpoint = "https://appleid.apple.com/auth/token"
)

type tokenExchange struct {
	identityToken oauth.SensitiveString
	tokenType     string
	expiresIn     int64
}

type tokenResponse struct {
	TokenType     string                `json:"token_type"`
	ExpiresIn     int64                 `json:"expires_in"`
	IdentityToken oauth.SensitiveString `json:"id_token"`
	Error         string                `json:"error"`
}

func (c *Client) exchange(ctx context.Context, clientID, code, redirectURI string) (tokenExchange, error) {
	clientID, err := c.clientID(clientID, "exchange code")
	if err != nil {
		return tokenExchange{}, err
	}
	if strings.TrimSpace(code) == "" {
		return tokenExchange{}, invalidInput("exchange code", errors.New("authorization code is required"))
	}
	if redirectURI != "" && strings.TrimSpace(redirectURI) == "" {
		return tokenExchange{}, invalidInput("exchange code", errors.New("redirect URI is invalid"))
	}
	secret, err := c.clientSecret(clientID)
	if err != nil {
		return tokenExchange{}, err
	}

	form := url.Values{
		"client_id":     {clientID},
		"client_secret": {secret.Value()},
		"code":          {code},
		"grant_type":    {"authorization_code"},
	}
	if redirectURI != "" {
		form.Set("redirect_uri", redirectURI)
	}
	req, err := http.NewRequest(http.MethodPost, tokenEndpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return tokenExchange{}, oauth.NewError(
			oauth.ProviderApple,
			oauth.ErrorKindInvalidConfig,
			"exchange code",
			err,
		)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	response, requestErr := c.do(ctx, "exchange code", req)
	if requestErr != nil && response == nil {
		return tokenExchange{}, requestErr
	}

	var payload tokenResponse
	if err := json.Unmarshal(response.Body, &payload); err != nil {
		if requestErr != nil {
			return tokenExchange{}, requestErr
		}
		return tokenExchange{}, oauth.NewError(
			oauth.ProviderApple,
			oauth.ErrorKindDecode,
			"exchange code",
			err,
		)
	}
	if payload.Error != "" {
		platformErr := oauth.NewError(
			oauth.ProviderApple,
			oauth.ErrorKindPlatform,
			"exchange code",
			requestErr,
		)
		platformErr.StatusCode = response.StatusCode
		platformErr.Code = payload.Error
		return tokenExchange{}, platformErr
	}
	if requestErr != nil {
		return tokenExchange{}, requestErr
	}
	if strings.TrimSpace(payload.IdentityToken.Value()) == "" {
		return tokenExchange{}, oauth.NewError(
			oauth.ProviderApple,
			oauth.ErrorKindInvalidResponse,
			"exchange code",
			errors.New("identity token is missing"),
		)
	}
	return tokenExchange{
		identityToken: payload.IdentityToken,
		tokenType:     payload.TokenType,
		expiresIn:     payload.ExpiresIn,
	}, nil
}
