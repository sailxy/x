package github_test

import (
	"context"
	"os"
	"time"

	"github.com/sailxy/x/oauth"
	"github.com/sailxy/x/oauth/github"
)

func ExampleClient_Authenticate() {
	client, err := github.New(github.Config{
		ClientID:     os.Getenv("GITHUB_CLIENT_ID"),
		ClientSecret: oauth.SensitiveString(os.Getenv("GITHUB_CLIENT_SECRET")),
	})
	if err != nil {
		return
	}

	// The backend creates, stores, and later consumes state exactly once.
	// The redirect URI used for the code exchange must match this value.
	const redirectURI = "https://example.com/oauth/github/callback"
	authorizationURL, err := client.AuthorizationURL(redirectURI, "short-lived-backend-state")
	if err != nil {
		return
	}
	_ = authorizationURL // Redirect the user's browser to this URL.

	// After the backend validates the callback state, authenticate with a
	// deadline so both GitHub requests have bounded execution time.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	result, err := client.Authenticate(ctx, "code-from-callback", redirectURI)
	if err != nil {
		return
	}

	// AccountID is the stable GitHub identity. Matching it to an application
	// user and issuing an application session remain backend responsibilities.
	accountID := result.User.AccountID
	_ = accountID

	// If the backend needs the access token for an immediate GitHub operation,
	// use it and discard it. Do not persist or log it.
	accessToken := result.Token.AccessToken.Value()
	_ = accessToken
}
