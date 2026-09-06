package github

import (
	"errors"
	"net/url"
	"testing"

	"github.com/sailxy/x/oauth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
