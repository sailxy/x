package github

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/sailxy/x/oauth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthenticate(t *testing.T) {
	const redirectURI = "https://example.com/oauth/github/callback"
	client := newLocalClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/login/oauth/access_token":
			require.NoError(t, r.ParseForm())
			assert.Equal(t, "authorization-code", r.Form.Get("code"))
			assert.Equal(t, redirectURI, r.Form.Get("redirect_uri"))
			_, _ = io.WriteString(w, `{"access_token":"github-access-token","scope":"read:user","token_type":"bearer"}`)
		case "/user":
			assertGitHubAPIRequest(t, r, "/user", "github-access-token")
			_, _ = io.WriteString(w, `{
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

	result, err := client.Authenticate(context.Background(), "authorization-code", redirectURI)
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
}

func TestNew(t *testing.T) {
	t.Run("valid default HTTP client", func(t *testing.T) {
		client, err := New(Config{
			ClientID:     "github-client-id",
			ClientSecret: oauth.SensitiveString("github-client-secret"),
		})
		require.NoError(t, err)
		require.NotNil(t, client)
		assert.NotNil(t, client.rest)
	})

	t.Run("valid injected HTTP client", func(t *testing.T) {
		client, err := New(Config{
			ClientID:     "github-client-id",
			ClientSecret: oauth.SensitiveString("github-client-secret"),
			HTTPClient:   &http.Client{Timeout: time.Second},
		})
		require.NoError(t, err)
		require.NotNil(t, client)
		assert.NotNil(t, client.rest)
	})

	for _, test := range []struct {
		name   string
		config Config
	}{
		{
			name: "missing client ID",
			config: Config{
				ClientSecret: oauth.SensitiveString("github-client-secret"),
			},
		},
		{
			name: "blank client ID",
			config: Config{
				ClientID:     "  ",
				ClientSecret: oauth.SensitiveString("github-client-secret"),
			},
		},
		{
			name: "missing client secret",
			config: Config{
				ClientID: "github-client-id",
			},
		},
		{
			name: "blank client secret",
			config: Config{
				ClientID:     "github-client-id",
				ClientSecret: oauth.SensitiveString("  "),
			},
		},
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
	config := Config{
		ClientID:     "github-client-id",
		ClientSecret: oauth.SensitiveString(secret),
		HTTPClient:   &http.Client{},
	}
	client, err := New(config)
	require.NoError(t, err)

	for _, value := range []any{config, client, *client} {
		for _, format := range []string{"%s", "%q", "%v", "%+v", "%#v"} {
			assert.NotContains(t, fmt.Sprintf(format, value), secret)
		}
	}
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
