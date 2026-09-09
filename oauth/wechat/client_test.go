package wechat

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthenticateAndUser(t *testing.T) {
	var userCalls int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/sns/oauth2/access_token":
			_, _ = w.Write([]byte(`{"access_token":"token","refresh_token":"refresh","expires_in":777,"scope":"snsapi_login","openid":"open","unionid":"union"}`))
		case "/sns/userinfo":
			userCalls++
			_, _ = w.Write([]byte(`{"openid":"open","unionid":"union","nickname":"nick","sex":1}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()
	target, err := url.Parse(server.URL)
	require.NoError(t, err)
	client, err := New(Config{AppID: "id", AppSecret: "secret", HTTPClient: &http.Client{Transport: rewrite{target: target, base: server.Client().Transport}}})
	require.NoError(t, err)
	result, err := client.Authenticate(context.Background(), "code")
	require.NoError(t, err)
	assert.Equal(t, "union", result.Identity.UnionID)
	assert.Zero(t, userCalls)
	user, err := client.User(context.Background(), result)
	require.NoError(t, err)
	assert.Equal(t, "nick", user.Nickname)
	assert.Equal(t, 1, userCalls)
}

func TestAuthorizationURL(t *testing.T) {
	c, err := New(Config{AppID: "id", AppSecret: "secret"})
	require.NoError(t, err)
	address, err := c.AuthorizationURL("https://example.com/callback", "state")
	require.NoError(t, err)
	assert.Contains(t, address, "#wechat_redirect")
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
