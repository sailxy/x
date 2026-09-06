package github

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/sailxy/x/oauth"
	"github.com/sailxy/x/rest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExchange(t *testing.T) {
	const redirectURI = "https://example.com/oauth/github/callback?next=%2Fitems"

	t.Run("success", func(t *testing.T) {
		client := newLocalClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodPost, r.Method)
			assert.Equal(t, "/login/oauth/access_token", r.URL.Path)
			assert.Empty(t, r.URL.RawQuery)
			assert.Equal(t, "application/json", r.Header.Get("Accept"))
			assert.Equal(t, "application/x-www-form-urlencoded", r.Header.Get("Content-Type"))
			assert.Equal(t, userAgent, r.Header.Get("User-Agent"))
			require.NoError(t, r.ParseForm())
			assert.Equal(t, "github-client-id", r.Form.Get("client_id"))
			assert.Equal(t, "github-client-secret", r.Form.Get("client_secret"))
			assert.Equal(t, "authorization-code", r.Form.Get("code"))
			assert.Equal(t, redirectURI, r.Form.Get("redirect_uri"))
			w.Header().Set("Content-Type", "application/json")
			_, _ = io.WriteString(w, `{"access_token":"github-access-token","scope":"","token_type":"bearer"}`)
		}))

		token, err := client.exchange(context.Background(), "authorization-code", redirectURI)
		require.NoError(t, err)
		assert.Equal(t, "github-access-token", token.AccessToken.Value())
		assert.Empty(t, token.Scope)
		assert.Equal(t, "bearer", token.TokenType)
	})

	t.Run("OAuth rejection with success HTTP status", func(t *testing.T) {
		client := newLocalClient(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			_, _ = io.WriteString(w, `{"error":"bad_verification_code","error_description":"authorization-code-must-not-leak"}`)
		}))

		token, err := client.exchange(context.Background(), "authorization-code-must-not-leak", redirectURI)
		assert.Zero(t, token)
		assert.ErrorIs(t, err, oauth.ErrPlatform)
		var oauthError *oauth.Error
		require.True(t, errors.As(err, &oauthError))
		assert.Equal(t, "bad_verification_code", oauthError.Code)
		assert.NotContains(t, err.Error(), "authorization-code-must-not-leak")
	})

	t.Run("non-success HTTP status", func(t *testing.T) {
		client := newLocalClient(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusBadGateway)
			_, _ = io.WriteString(w, `{"message":"upstream failed"}`)
		}))

		token, err := client.exchange(context.Background(), "authorization-code", redirectURI)
		assert.Zero(t, token)
		assert.ErrorIs(t, err, oauth.ErrPlatform)
		var oauthError *oauth.Error
		require.True(t, errors.As(err, &oauthError))
		assert.Equal(t, http.StatusBadGateway, oauthError.StatusCode)
	})

	for _, test := range []struct {
		name   string
		body   string
		target error
	}{
		{name: "malformed JSON", body: `{`, target: oauth.ErrDecode},
		{name: "missing access token", body: `{}`, target: oauth.ErrInvalidResponse},
	} {
		t.Run(test.name, func(t *testing.T) {
			client := newLocalClient(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				_, _ = io.WriteString(w, test.body)
			}))
			token, err := client.exchange(context.Background(), "authorization-code", redirectURI)
			assert.Zero(t, token)
			assert.ErrorIs(t, err, test.target)
		})
	}

	t.Run("oversized response", func(t *testing.T) {
		client := newLocalClient(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			_, _ = io.WriteString(w, strings.Repeat("x", int(rest.MaxResponseBodyBytes)+1))
		}))
		token, err := client.exchange(context.Background(), "authorization-code", redirectURI)
		assert.Zero(t, token)
		assert.ErrorIs(t, err, oauth.ErrInvalidResponse)
	})

	for _, test := range []struct {
		name        string
		code        string
		redirectURI string
	}{
		{name: "missing code", redirectURI: redirectURI},
		{name: "blank code", code: "  ", redirectURI: redirectURI},
		{name: "missing redirect URI", code: "authorization-code"},
		{name: "blank redirect URI", code: "authorization-code", redirectURI: "  "},
	} {
		t.Run(test.name, func(t *testing.T) {
			client := newWaitingClient(t)
			token, err := client.exchange(context.Background(), test.code, test.redirectURI)
			assert.Zero(t, token)
			assert.ErrorIs(t, err, oauth.ErrInvalidInput)
		})
	}
}

func TestExchangeCancellation(t *testing.T) {
	const redirectURI = "https://example.com/callback"

	t.Run("canceled", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		token, err := newWaitingClient(t).exchange(ctx, "authorization-code", redirectURI)
		assert.Zero(t, token)
		assert.ErrorIs(t, err, oauth.ErrTransport)
		assert.ErrorIs(t, err, context.Canceled)
	})

	t.Run("deadline", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
		defer cancel()
		token, err := newWaitingClient(t).exchange(ctx, "authorization-code", redirectURI)
		assert.Zero(t, token)
		assert.ErrorIs(t, err, oauth.ErrTransport)
		assert.ErrorIs(t, err, context.DeadlineExceeded)
	})
}
