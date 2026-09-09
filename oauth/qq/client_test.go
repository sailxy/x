package qq

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/sailxy/x/oauth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	client, err := New(Config{
		ClientID:     "qq-client-id",
		ClientSecret: "qq-client-secret",
		HTTPClient:   &http.Client{},
	})
	require.NoError(t, err)
	assert.NotNil(t, client)
	assert.NotNil(t, client.rest)

	for _, config := range []Config{
		{ClientSecret: "qq-client-secret"},
		{ClientID: "qq-client-id"},
	} {
		client, err := New(config)
		assert.Nil(t, client)
		assert.ErrorIs(t, err, oauth.ErrInvalidConfig)
		var oauthError *oauth.Error
		require.True(t, errors.As(err, &oauthError))
		assert.Equal(t, oauth.ProviderQQ, oauthError.Provider)
	}
}

func TestAuthorizationURL(t *testing.T) {
	client := testClient(t)
	address, err := client.AuthorizationURL("https://example.com/callback?next=/account", "state + / ? & =")
	require.NoError(t, err)

	parsed, err := url.Parse(address)
	require.NoError(t, err)
	assert.Equal(t, "https", parsed.Scheme)
	assert.Equal(t, "graph.qq.com", parsed.Host)
	assert.Equal(t, "/oauth2.0/authorize", parsed.Path)
	assert.Equal(t, url.Values{
		"response_type": {"code"},
		"client_id":     {"qq-client-id"},
		"redirect_uri":  {"https://example.com/callback?next=/account"},
		"state":         {"state + / ? & ="},
	}, parsed.Query())
}

func TestAuthorizationURLRejectsInvalidInput(t *testing.T) {
	client := testClient(t)
	for _, request := range [][2]string{{"", "state"}, {"https://example.com/callback", ""}} {
		address, err := client.AuthorizationURL(request[0], request[1])
		assert.Empty(t, address)
		assert.ErrorIs(t, err, oauth.ErrInvalidInput)
	}
}

func TestAuthenticateAndUser(t *testing.T) {
	var userCalls int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/oauth2.0/token":
			_, _ = w.Write([]byte(`{"access_token":"token","refresh_token":"refresh","expires_in":"777"}`))
		case "/oauth2.0/me":
			_, _ = w.Write([]byte(`{"openid":"open","unionid":"union"}`))
		case "/user/get_user_info":
			userCalls++
			_, _ = w.Write([]byte(`{"ret":0,"nickname":"nick","figureurl_qq":"avatar"}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()
	target, err := url.Parse(server.URL)
	require.NoError(t, err)
	client, err := New(Config{ClientID: "id", ClientSecret: "secret", HTTPClient: &http.Client{Transport: rewrite{target: target, base: server.Client().Transport}}})
	require.NoError(t, err)
	result, err := client.Authenticate(context.Background(), "code", "https://example.com/callback")
	require.NoError(t, err)
	assert.Equal(t, "union", result.Identity.UnionID)
	assert.Equal(t, "token", result.Token.AccessToken.Value())
	assert.Zero(t, userCalls)
	user, err := client.User(context.Background(), result)
	require.NoError(t, err)
	assert.Equal(t, "nick", user.Nickname)
	assert.Equal(t, 1, userCalls)
}

func testClient(t *testing.T) *Client {
	t.Helper()
	client, err := New(Config{ClientID: "qq-client-id", ClientSecret: "qq-client-secret"})
	require.NoError(t, err)
	return client
}

type rewrite struct {
	target *url.URL
	base   http.RoundTripper
}

func (r rewrite) RoundTrip(req *http.Request) (*http.Response, error) {
	clone := req.Clone(req.Context())
	u := *req.URL
	u.Scheme, u.Host = r.target.Scheme, r.target.Host
	clone.URL, clone.Host = &u, ""
	return r.base.RoundTrip(clone)
}
