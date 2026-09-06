package apple

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/sailxy/x/oauth"
	"github.com/sailxy/x/rest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExchange(t *testing.T) {
	const redirectURI = "https://example.com/oauth/apple/callback?next=%2Faccount"

	t.Run("success", func(t *testing.T) {
		client := newLocalAppleClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodPost, r.Method)
			assert.Equal(t, "/auth/token", r.URL.Path)
			assert.Empty(t, r.URL.RawQuery)
			assert.Equal(t, "application/json", r.Header.Get("Accept"))
			assert.Equal(t, "application/x-www-form-urlencoded", r.Header.Get("Content-Type"))
			require.NoError(t, r.ParseForm())
			assert.Equal(t, "com.example.web", r.Form.Get("client_id"))
			assert.NotEmpty(t, r.Form.Get("client_secret"))
			assert.Equal(t, "authorization-code", r.Form.Get("code"))
			assert.Equal(t, "authorization_code", r.Form.Get("grant_type"))
			assert.Equal(t, redirectURI, r.Form.Get("redirect_uri"))
			_, _ = io.WriteString(w, `{
				"access_token":"apple-access-token",
				"token_type":"Bearer",
				"expires_in":3600,
				"refresh_token":"apple-refresh-token",
				"id_token":"apple-identity-token"
			}`)
		}))

		result, err := client.exchange(context.Background(), "com.example.web", "authorization-code", redirectURI)
		require.NoError(t, err)
		assert.Equal(t, "apple-identity-token", result.identityToken.Value())
		assert.Equal(t, "Bearer", result.tokenType)
		assert.Equal(t, int64(3600), result.expiresIn)
	})

	t.Run("redirect URI omitted", func(t *testing.T) {
		client := newLocalAppleClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.NoError(t, r.ParseForm())
			assert.NotContains(t, r.Form, "redirect_uri")
			_, _ = io.WriteString(w, `{"id_token":"apple-identity-token"}`)
		}))

		result, err := client.exchange(context.Background(), "com.example.ios", "authorization-code", "")
		require.NoError(t, err)
		assert.Equal(t, "apple-identity-token", result.identityToken.Value())
	})
}

func TestExchangeRejectsAppleFailures(t *testing.T) {
	t.Run("platform rejection", func(t *testing.T) {
		var sentSecret string
		client := newLocalAppleClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.NoError(t, r.ParseForm())
			sentSecret = r.Form.Get("client_secret")
			w.WriteHeader(http.StatusBadRequest)
			_, _ = io.WriteString(w, `{"error":"invalid_grant","error_description":"authorization-code-must-not-leak"}`)
		}))

		result, err := client.exchange(context.Background(), "com.example.web", "authorization-code-must-not-leak", "")
		assert.Zero(t, result)
		assert.ErrorIs(t, err, oauth.ErrPlatform)
		var oauthError *oauth.Error
		require.True(t, errors.As(err, &oauthError))
		assert.Equal(t, "invalid_grant", oauthError.Code)
		assert.Equal(t, http.StatusBadRequest, oauthError.StatusCode)
		assert.NotContains(t, err.Error(), "authorization-code-must-not-leak")
		assert.NotEmpty(t, sentSecret)
		assert.NotContains(t, err.Error(), sentSecret)
	})

	t.Run("non-success HTTP status", func(t *testing.T) {
		client := newLocalAppleClient(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusBadGateway)
			_, _ = io.WriteString(w, `{}`)
		}))

		result, err := client.exchange(context.Background(), "com.example.web", "authorization-code", "")
		assert.Zero(t, result)
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
		{name: "missing identity token", body: `{"access_token":"must-not-be-returned"}`, target: oauth.ErrInvalidResponse},
	} {
		t.Run(test.name, func(t *testing.T) {
			client := newLocalAppleClient(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				_, _ = io.WriteString(w, test.body)
			}))

			result, err := client.exchange(context.Background(), "com.example.web", "authorization-code", "")
			assert.Zero(t, result)
			assert.ErrorIs(t, err, test.target)
			assert.NotContains(t, err.Error(), "must-not-be-returned")
		})
	}

	t.Run("oversized response", func(t *testing.T) {
		client := newLocalAppleClient(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			_, _ = io.WriteString(w, strings.Repeat("x", int(rest.MaxResponseBodyBytes)+1))
		}))

		result, err := client.exchange(context.Background(), "com.example.web", "authorization-code", "")
		assert.Zero(t, result)
		assert.ErrorIs(t, err, oauth.ErrInvalidResponse)
	})
}

func TestExchangeRejectsInvalidInput(t *testing.T) {
	client := newAppleClientWithTransport(t, roundTripperFunc(func(*http.Request) (*http.Response, error) {
		t.Fatal("exchange sent a request for invalid input")
		return nil, nil
	}))

	for _, test := range []struct {
		name        string
		clientID    string
		code        string
		redirectURI string
	}{
		{name: "missing client ID", code: "authorization-code"},
		{name: "unknown client ID", clientID: "com.example.unknown", code: "authorization-code"},
		{name: "missing code", clientID: "com.example.web"},
		{name: "blank code", clientID: "com.example.web", code: "  "},
		{name: "blank redirect URI", clientID: "com.example.web", code: "authorization-code", redirectURI: "  "},
	} {
		t.Run(test.name, func(t *testing.T) {
			result, err := client.exchange(context.Background(), test.clientID, test.code, test.redirectURI)
			assert.Zero(t, result)
			assert.ErrorIs(t, err, oauth.ErrInvalidInput)
		})
	}
}

func TestExchangeCancellation(t *testing.T) {
	client := newAppleClientWithTransport(t, waitForAppleContextTransport{})

	t.Run("canceled", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		result, err := client.exchange(ctx, "com.example.web", "authorization-code", "")
		assert.Zero(t, result)
		assert.ErrorIs(t, err, oauth.ErrTransport)
		assert.ErrorIs(t, err, context.Canceled)
	})

	t.Run("deadline", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
		defer cancel()
		result, err := client.exchange(ctx, "com.example.web", "authorization-code", "")
		assert.Zero(t, result)
		assert.ErrorIs(t, err, oauth.ErrTransport)
		assert.ErrorIs(t, err, context.DeadlineExceeded)
	})

	t.Run("nil context", func(t *testing.T) {
		result, err := client.exchange(nil, "com.example.web", "authorization-code", "")
		assert.Zero(t, result)
		assert.ErrorIs(t, err, oauth.ErrInvalidInput)
	})
}

func newLocalAppleClient(t *testing.T, handler http.Handler) *Client {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	target, err := url.Parse(server.URL)
	require.NoError(t, err)
	return newAppleClientWithTransport(t, appleRewriteTransport{
		target: target,
		base:   server.Client().Transport,
	})
}

func newAppleClientWithTransport(t *testing.T, transport http.RoundTripper) *Client {
	t.Helper()
	client, err := New(Config{
		TeamID:        "TEAM123456",
		ClientIDs:     []string{"com.example.web", "com.example.ios"},
		KeyID:         "KEY1234567",
		PrivateKeyPEM: testPKCS8PEM(t, fixedTestPrivateKey()),
		HTTPClient:    &http.Client{Transport: transport},
		Now: func() time.Time {
			return time.Date(2026, time.September, 7, 12, 0, 0, 0, time.UTC)
		},
	})
	require.NoError(t, err)
	return client
}

type appleRewriteTransport struct {
	target *url.URL
	base   http.RoundTripper
}

func (t appleRewriteTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	clone := req.Clone(req.Context())
	requestURL := *req.URL
	requestURL.Scheme = t.target.Scheme
	requestURL.Host = t.target.Host
	clone.URL = &requestURL
	clone.Host = ""
	return t.base.RoundTrip(clone)
}

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

type waitForAppleContextTransport struct{}

func (waitForAppleContextTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	<-req.Context().Done()
	return nil, req.Context().Err()
}
