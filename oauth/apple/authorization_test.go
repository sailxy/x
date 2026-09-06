package apple

import (
	"crypto/elliptic"
	"errors"
	"net/url"
	"testing"

	"github.com/sailxy/x/oauth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthorizationURL(t *testing.T) {
	client := testAppleClient(t)

	address, err := client.AuthorizationURL(AuthorizationRequest{
		ClientID:    "com.example.web",
		RedirectURI: "https://example.com/callback?from=apple&next=/account",
		State:       "state + / ? & =",
		Nonce:       "nonce + / ? & =",
		Scopes:      []string{"name", "email", "name"},
	})
	require.NoError(t, err)

	parsed, err := url.Parse(address)
	require.NoError(t, err)
	assert.Equal(t, "https", parsed.Scheme)
	assert.Equal(t, "appleid.apple.com", parsed.Host)
	assert.Equal(t, "/auth/authorize", parsed.Path)
	assert.Equal(t, url.Values{
		"client_id":     {"com.example.web"},
		"redirect_uri":  {"https://example.com/callback?from=apple&next=/account"},
		"response_type": {"code"},
		"response_mode": {"form_post"},
		"state":         {"state + / ? & ="},
		"nonce":         {"nonce + / ? & ="},
		"scope":         {"name email"},
	}, parsed.Query())
}

func TestAuthorizationURLWithoutOptionalValues(t *testing.T) {
	client := testAppleClient(t)

	address, err := client.AuthorizationURL(AuthorizationRequest{
		ClientID:    "com.example.web",
		RedirectURI: "https://example.com/callback",
		State:       "state",
	})
	require.NoError(t, err)

	parsed, err := url.Parse(address)
	require.NoError(t, err)
	assert.NotContains(t, parsed.Query(), "nonce")
	assert.NotContains(t, parsed.Query(), "scope")
	assert.Equal(t, "form_post", parsed.Query().Get("response_mode"))
}

func TestAuthorizationURLRejectsInvalidInput(t *testing.T) {
	client := testAppleClient(t)
	valid := func() AuthorizationRequest {
		return AuthorizationRequest{
			ClientID:    "com.example.web",
			RedirectURI: "https://example.com/callback",
			State:       "state",
		}
	}
	tests := []struct {
		name   string
		mutate func(*AuthorizationRequest)
	}{
		{name: "missing client ID", mutate: func(r *AuthorizationRequest) { r.ClientID = "" }},
		{name: "unknown client ID", mutate: func(r *AuthorizationRequest) { r.ClientID = "com.example.unknown" }},
		{name: "missing redirect URI", mutate: func(r *AuthorizationRequest) { r.RedirectURI = "  " }},
		{name: "missing state", mutate: func(r *AuthorizationRequest) { r.State = "" }},
		{name: "unsupported scope", mutate: func(r *AuthorizationRequest) { r.Scopes = []string{"openid"} }},
		{name: "blank scope", mutate: func(r *AuthorizationRequest) { r.Scopes = []string{"email", " "} }},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := valid()
			test.mutate(&request)
			address, err := client.AuthorizationURL(request)
			assert.Empty(t, address)
			assert.ErrorIs(t, err, oauth.ErrInvalidInput)

			var oauthError *oauth.Error
			require.True(t, errors.As(err, &oauthError))
			assert.Equal(t, oauth.ProviderApple, oauthError.Provider)
			assert.Equal(t, "create authorization URL", oauthError.Operation)
		})
	}
}

func testAppleClient(t *testing.T) *Client {
	t.Helper()
	client, err := New(Config{
		TeamID:        "TEAM123456",
		ClientIDs:     []string{"com.example.web", "com.example.ios"},
		KeyID:         "KEY1234567",
		PrivateKeyPEM: testPKCS8PEM(t, testECPrivateKey(t, elliptic.P256())),
	})
	require.NoError(t, err)
	return client
}
