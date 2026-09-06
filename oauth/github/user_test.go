package github

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"testing"

	"github.com/sailxy/x/oauth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetCurrentUser(t *testing.T) {
	const rawToken = "github-user-access-token"

	t.Run("success with available public profile", func(t *testing.T) {
		client := newLocalClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assertGitHubAPIRequest(t, r, "/user", rawToken)
			_, _ = io.WriteString(w, `{
				"id": 123456789,
				"node_id": "MDQ6VXNlcjE=",
				"login": "octocat",
				"name": "The Octocat",
				"avatar_url": "https://avatars.example/octocat",
				"email": "octocat@example.com"
			}`)
		}))

		user, err := client.getCurrentUser(context.Background(), oauth.SensitiveString(rawToken))
		require.NoError(t, err)
		assert.Equal(t, int64(123456789), user.AccountID)
		assert.Equal(t, "MDQ6VXNlcjE=", user.NodeID)
		assert.Equal(t, "octocat", user.Login)
		assert.Equal(t, "The Octocat", user.Name)
		assert.Equal(t, "https://avatars.example/octocat", user.AvatarURL)
		assert.Equal(t, "octocat@example.com", user.Email)
	})

	t.Run("missing public email does not fail identity", func(t *testing.T) {
		client := newLocalClient(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			_, _ = io.WriteString(w, `{"id":123456789,"login":"octocat","email":null}`)
		}))

		user, err := client.getCurrentUser(context.Background(), oauth.SensitiveString(rawToken))
		require.NoError(t, err)
		assert.Equal(t, int64(123456789), user.AccountID)
		assert.Empty(t, user.Email)
	})

	for _, test := range []struct {
		name   string
		status int
		body   string
		target error
	}{
		{name: "non-success", status: http.StatusUnauthorized, body: `{"message":"bad credentials"}`, target: oauth.ErrPlatform},
		{name: "malformed JSON", status: http.StatusOK, body: `{`, target: oauth.ErrDecode},
		{name: "missing account ID", status: http.StatusOK, body: `{"login":"octocat"}`, target: oauth.ErrInvalidResponse},
		{name: "negative account ID", status: http.StatusOK, body: `{"id":-1}`, target: oauth.ErrInvalidResponse},
	} {
		t.Run(test.name, func(t *testing.T) {
			client := newLocalClient(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(test.status)
				_, _ = io.WriteString(w, test.body)
			}))
			user, err := client.getCurrentUser(context.Background(), oauth.SensitiveString(rawToken))
			assert.Zero(t, user)
			assert.ErrorIs(t, err, test.target)
		})
	}

	t.Run("missing access token", func(t *testing.T) {
		client := newWaitingClient(t)
		user, err := client.getCurrentUser(context.Background(), "")
		assert.Zero(t, user)
		assert.ErrorIs(t, err, oauth.ErrInvalidInput)
	})
}

func assertGitHubAPIRequest(t *testing.T, req *http.Request, path, token string) {
	t.Helper()
	assert.Equal(t, http.MethodGet, req.Method)
	assert.Equal(t, path, req.URL.Path)
	assert.Equal(t, "Bearer "+token, req.Header.Get("Authorization"))
	assert.Equal(t, apiAcceptHeader, req.Header.Get("Accept"))
	assert.Equal(t, apiVersion, req.Header.Get("X-GitHub-Api-Version"))
	assert.Equal(t, userAgent, req.Header.Get("User-Agent"))
	assert.NotContains(t, req.URL.String(), token)
	_, err := url.ParseRequestURI(req.URL.RequestURI())
	require.NoError(t, err)
}
