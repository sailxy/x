package github

import (
	"context"
	"errors"
	"fmt"
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

const testRedirectURI = "https://example.com/oauth/github/callback?next=%2Fitems"

func TestAuthorizationURL(t *testing.T) {
	client, err := New(Config{
		ClientID:     "github-client-id",
		ClientSecret: oauth.SensitiveString("github-client-secret"),
	})
	require.NoError(t, err)

	t.Run("required parameters without extra scope", func(t *testing.T) {
		got, err := client.AuthorizationURL("https://example.com/oauth/github/callback", "state-123")
		require.NoError(t, err)

		parsed, err := url.Parse(got)
		require.NoError(t, err)
		assert.Equal(t, "https", parsed.Scheme)
		assert.Equal(t, "github.com", parsed.Host)
		assert.Equal(t, "/login/oauth/authorize", parsed.Path)
		assert.Equal(t, "github-client-id", parsed.Query().Get("client_id"))
		assert.Equal(t, "https://example.com/oauth/github/callback", parsed.Query().Get("redirect_uri"))
		assert.Empty(t, parsed.Query().Get("scope"))
		assert.Equal(t, "state-123", parsed.Query().Get("state"))
		assert.Len(t, parsed.Query(), 3)
	})

	t.Run("reserved characters round trip", func(t *testing.T) {
		redirectURI := "https://example.com/callback?next=/items?a=1&name=张 三#result"
		state := "state +/=?&中文"
		got, err := client.AuthorizationURL(redirectURI, state)
		require.NoError(t, err)

		parsed, err := url.Parse(got)
		require.NoError(t, err)
		assert.Equal(t, redirectURI, parsed.Query().Get("redirect_uri"))
		assert.Equal(t, state, parsed.Query().Get("state"))
		assert.NotContains(t, got, "张 三")
	})

	for _, test := range []struct {
		name        string
		redirectURI string
		state       string
	}{
		{name: "missing redirect URI", state: "state-123"},
		{name: "blank redirect URI", redirectURI: "  ", state: "state-123"},
		{name: "missing state", redirectURI: "https://example.com/callback"},
		{name: "blank state", redirectURI: "https://example.com/callback", state: "  "},
	} {
		t.Run(test.name, func(t *testing.T) {
			got, err := client.AuthorizationURL(test.redirectURI, test.state)
			assert.Empty(t, got)
			assert.ErrorIs(t, err, oauth.ErrInvalidInput)

			var oauthError *oauth.Error
			require.True(t, errors.As(err, &oauthError))
			assert.Equal(t, oauth.ProviderGitHub, oauthError.Provider)
			assert.Equal(t, "build authorization URL", oauthError.Operation)
		})
	}
}

func TestAuthenticate(t *testing.T) {
	t.Run("returns token and user", func(t *testing.T) {
		client := newLocalClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/login/oauth/access_token":
				assertTokenRequest(t, r, "authorization-code", testRedirectURI)
				writeResponse(t, w, http.StatusOK, `{"access_token":"github-access-token","scope":"read:user","token_type":"bearer"}`)
			case "/user":
				assertGitHubAPIRequest(t, r, "github-access-token")
				writeResponse(t, w, http.StatusOK, `{
					"id":123456789,
					"node_id":"MDQ6VXNlcjE=",
					"login":"octocat",
					"name":"The Octocat",
					"avatar_url":"https://avatars.example/octocat",
					"email":"octocat@example.com"
				}`)
			default:
				http.NotFound(w, r)
			}
		}))

		result, err := client.Authenticate(context.Background(), "authorization-code", testRedirectURI)
		require.NoError(t, err)
		assert.Equal(t, "github-access-token", result.Token.AccessToken.Value())
		assert.Equal(t, "read:user", result.Token.Scope)
		assert.Equal(t, "bearer", result.Token.TokenType)
		assert.Equal(t, int64(123456789), result.User.AccountID)
		assert.Equal(t, "MDQ6VXNlcjE=", result.User.NodeID)
		assert.Equal(t, "octocat", result.User.Login)
		assert.Equal(t, "The Octocat", result.User.Name)
		assert.Equal(t, "https://avatars.example/octocat", result.User.AvatarURL)
		assert.Equal(t, "octocat@example.com", result.User.Email)
	})

	t.Run("public email is optional", func(t *testing.T) {
		client := newFlowClient(t,
			responseFixture{body: `{"access_token":"github-access-token"}`},
			responseFixture{body: `{"id":123456789,"login":"octocat","email":null}`},
		)

		result, err := client.Authenticate(context.Background(), "authorization-code", testRedirectURI)
		require.NoError(t, err)
		assert.Equal(t, int64(123456789), result.User.AccountID)
		assert.Empty(t, result.User.Email)
	})
}

func TestAuthenticateRejectsTokenFailures(t *testing.T) {
	t.Run("OAuth rejection", func(t *testing.T) {
		client := newFlowClient(t,
			responseFixture{body: `{"error":"bad_verification_code","error_description":"authorization-code-must-not-leak"}`},
			responseFixture{},
		)

		result, err := client.Authenticate(context.Background(), "authorization-code-must-not-leak", testRedirectURI)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, oauth.ErrPlatform)
		var oauthError *oauth.Error
		require.True(t, errors.As(err, &oauthError))
		assert.Equal(t, "bad_verification_code", oauthError.Code)
		assert.NotContains(t, err.Error(), "authorization-code-must-not-leak")
	})

	t.Run("non-success status", func(t *testing.T) {
		client := newFlowClient(t,
			responseFixture{status: http.StatusBadGateway, body: `{"message":"upstream failed"}`},
			responseFixture{},
		)

		result, err := client.Authenticate(context.Background(), "authorization-code", testRedirectURI)
		assert.Nil(t, result)
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
			client := newFlowClient(t, responseFixture{body: test.body}, responseFixture{})
			result, err := client.Authenticate(context.Background(), "authorization-code", testRedirectURI)
			assert.Nil(t, result)
			assert.ErrorIs(t, err, test.target)
		})
	}

	t.Run("oversized response", func(t *testing.T) {
		client := newFlowClient(t,
			responseFixture{body: strings.Repeat("x", int(rest.MaxResponseBodyBytes)+1)},
			responseFixture{},
		)
		result, err := client.Authenticate(context.Background(), "authorization-code", testRedirectURI)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, oauth.ErrInvalidResponse)
	})
}

func TestAuthenticateRejectsUserFailures(t *testing.T) {
	for _, test := range []struct {
		name   string
		status int
		body   string
		target error
	}{
		{name: "non-success status", status: http.StatusUnauthorized, body: `{"message":"bad credentials"}`, target: oauth.ErrPlatform},
		{name: "malformed JSON", body: `{`, target: oauth.ErrDecode},
		{name: "missing account ID", body: `{"login":"octocat"}`, target: oauth.ErrInvalidResponse},
		{name: "negative account ID", body: `{"id":-1}`, target: oauth.ErrInvalidResponse},
	} {
		t.Run(test.name, func(t *testing.T) {
			client := newFlowClient(t,
				responseFixture{body: `{"access_token":"github-access-token"}`},
				responseFixture{status: test.status, body: test.body},
			)
			result, err := client.Authenticate(context.Background(), "authorization-code", testRedirectURI)
			assert.Nil(t, result)
			assert.ErrorIs(t, err, test.target)
		})
	}
}

func TestAuthenticateRejectsInvalidInput(t *testing.T) {
	for _, test := range []struct {
		name        string
		code        string
		redirectURI string
	}{
		{name: "missing code", redirectURI: testRedirectURI},
		{name: "blank code", code: "  ", redirectURI: testRedirectURI},
		{name: "missing redirect URI", code: "authorization-code"},
		{name: "blank redirect URI", code: "authorization-code", redirectURI: "  "},
	} {
		t.Run(test.name, func(t *testing.T) {
			result, err := newWaitingClient(t).Authenticate(context.Background(), test.code, test.redirectURI)
			assert.Nil(t, result)
			assert.ErrorIs(t, err, oauth.ErrInvalidInput)
		})
	}
}

func TestAuthenticateRespectsContext(t *testing.T) {
	t.Run("canceled", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		result, err := newWaitingClient(t).Authenticate(ctx, "authorization-code", testRedirectURI)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, oauth.ErrTransport)
		assert.ErrorIs(t, err, context.Canceled)
	})

	t.Run("deadline", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
		defer cancel()
		result, err := newWaitingClient(t).Authenticate(ctx, "authorization-code", testRedirectURI)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, oauth.ErrTransport)
		assert.ErrorIs(t, err, context.DeadlineExceeded)
	})
}

func TestNew(t *testing.T) {
	t.Run("valid default HTTP client", func(t *testing.T) {
		client, err := New(Config{ClientID: "github-client-id", ClientSecret: oauth.SensitiveString("github-client-secret")})
		require.NoError(t, err)
		assert.NotNil(t, client.rest)
	})

	t.Run("valid injected HTTP client", func(t *testing.T) {
		client, err := New(Config{
			ClientID: "github-client-id", ClientSecret: oauth.SensitiveString("github-client-secret"),
			HTTPClient: &http.Client{Timeout: time.Second},
		})
		require.NoError(t, err)
		assert.NotNil(t, client.rest)
	})

	for _, test := range []struct {
		name   string
		config Config
	}{
		{name: "missing client ID", config: Config{ClientSecret: oauth.SensitiveString("github-client-secret")}},
		{name: "blank client ID", config: Config{ClientID: "  ", ClientSecret: oauth.SensitiveString("github-client-secret")}},
		{name: "missing client secret", config: Config{ClientID: "github-client-id"}},
		{name: "blank client secret", config: Config{ClientID: "github-client-id", ClientSecret: oauth.SensitiveString("  ")}},
	} {
		t.Run(test.name, func(t *testing.T) {
			client, err := New(test.config)
			assert.Nil(t, client)
			assert.ErrorIs(t, err, oauth.ErrInvalidConfig)
			var oauthError *oauth.Error
			require.True(t, errors.As(err, &oauthError))
			assert.Equal(t, oauth.ProviderGitHub, oauthError.Provider)
		})
	}
}

func TestFormattingOmitsGitHubCredentials(t *testing.T) {
	const secret = "github-client-secret-fixture"
	config := Config{ClientID: "github-client-id", ClientSecret: oauth.SensitiveString(secret), HTTPClient: &http.Client{}}
	client, err := New(config)
	require.NoError(t, err)

	for _, value := range []any{config, client, *client} {
		for _, format := range []string{"%s", "%q", "%v", "%+v", "%#v"} {
			assert.NotContains(t, fmt.Sprintf(format, value), secret)
		}
	}
}

type responseFixture struct {
	status int
	body   string
}

func newFlowClient(t *testing.T, token, user responseFixture) *Client {
	t.Helper()
	return newLocalClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/login/oauth/access_token":
			writeResponse(t, w, token.status, token.body)
		case "/user":
			assertGitHubAPIRequest(t, r, "github-access-token")
			writeResponse(t, w, user.status, user.body)
		default:
			http.NotFound(w, r)
		}
	}))
}

func writeResponse(t *testing.T, w http.ResponseWriter, status int, body string) {
	t.Helper()
	if status == 0 {
		status = http.StatusOK
	}
	w.WriteHeader(status)
	_, err := io.WriteString(w, body)
	require.NoError(t, err)
}

func assertTokenRequest(t *testing.T, r *http.Request, code, redirectURI string) {
	t.Helper()
	assert.Equal(t, http.MethodPost, r.Method)
	assert.Equal(t, "/login/oauth/access_token", r.URL.Path)
	assert.Empty(t, r.URL.RawQuery)
	assert.Equal(t, "application/json", r.Header.Get("Accept"))
	assert.Equal(t, "application/x-www-form-urlencoded", r.Header.Get("Content-Type"))
	assert.Equal(t, userAgent, r.Header.Get("User-Agent"))
	require.NoError(t, r.ParseForm())
	assert.Equal(t, "github-client-id", r.Form.Get("client_id"))
	assert.Equal(t, "github-client-secret", r.Form.Get("client_secret"))
	assert.Equal(t, code, r.Form.Get("code"))
	assert.Equal(t, redirectURI, r.Form.Get("redirect_uri"))
}

func assertGitHubAPIRequest(t *testing.T, req *http.Request, token string) {
	t.Helper()
	assert.Equal(t, http.MethodGet, req.Method)
	assert.Equal(t, "/user", req.URL.Path)
	assert.Equal(t, "Bearer "+token, req.Header.Get("Authorization"))
	assert.Equal(t, apiAcceptHeader, req.Header.Get("Accept"))
	assert.Equal(t, apiVersion, req.Header.Get("X-GitHub-Api-Version"))
	assert.Equal(t, userAgent, req.Header.Get("User-Agent"))
	assert.NotContains(t, req.URL.String(), token)
	_, err := url.ParseRequestURI(req.URL.RequestURI())
	require.NoError(t, err)
}

func newLocalClient(t *testing.T, handler http.Handler) *Client {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	target, err := url.Parse(server.URL)
	require.NoError(t, err)
	client, err := New(Config{
		ClientID:     "github-client-id",
		ClientSecret: oauth.SensitiveString("github-client-secret"),
		HTTPClient: &http.Client{Transport: rewriteTransport{
			target: target,
			base:   server.Client().Transport,
		}},
	})
	require.NoError(t, err)
	return client
}

type rewriteTransport struct {
	target *url.URL
	base   http.RoundTripper
}

func (t rewriteTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	clone := req.Clone(req.Context())
	requestURL := *req.URL
	requestURL.Scheme = t.target.Scheme
	requestURL.Host = t.target.Host
	clone.URL = &requestURL
	clone.Host = ""
	return t.base.RoundTrip(clone)
}

type waitForContextTransport struct{}

func (waitForContextTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	<-req.Context().Done()
	return nil, req.Context().Err()
}

func newWaitingClient(t *testing.T) *Client {
	t.Helper()
	client, err := New(Config{
		ClientID:     "github-client-id",
		ClientSecret: oauth.SensitiveString("github-client-secret"),
		HTTPClient:   &http.Client{Transport: waitForContextTransport{}},
	})
	require.NoError(t, err)
	return client
}
