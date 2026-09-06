package apple_test

import (
	"context"
	"os"
	"time"

	"github.com/sailxy/x/oauth"
	"github.com/sailxy/x/oauth/apple"
)

func ExampleClient_Authenticate() {
	privateKey, err := os.ReadFile(os.Getenv("APPLE_PRIVATE_KEY_PATH"))
	if err != nil {
		return
	}
	clientID := os.Getenv("APPLE_WEB_CLIENT_ID")
	client, err := apple.New(apple.Config{
		TeamID:        os.Getenv("APPLE_TEAM_ID"),
		ClientIDs:     []string{clientID},
		KeyID:         os.Getenv("APPLE_KEY_ID"),
		PrivateKeyPEM: oauth.SensitiveString(privateKey),
	})
	if err != nil {
		return
	}

	// ClientID is selected by the backend from its configuration, never from
	// an arbitrary audience supplied by the client. State and nonce are stored
	// by the backend and consumed with the callback.
	const redirectURI = "https://example.com/oauth/apple/callback"
	const expectedNonceClaim = "nonce-associated-with-backend-state"
	authorizationURL, err := client.AuthorizationURL(apple.AuthorizationRequest{
		ClientID:    clientID,
		RedirectURI: redirectURI,
		State:       "short-lived-backend-state",
		Nonce:       expectedNonceClaim,
		Scopes:      []string{"name", "email"},
	})
	if err != nil {
		return
	}
	_ = authorizationURL // Redirect the user's browser to this URL.

	// After validating state, use the same Client ID and redirect URI for the
	// code exchange. Native clients that hash a raw nonce pass the hash expected
	// in the identity-token claim as ExpectedNonceClaim.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	result, err := client.Authenticate(ctx, apple.AuthenticateRequest{
		ClientID:           clientID,
		Code:               "code-from-callback",
		RedirectURI:        redirectURI,
		ExpectedNonceClaim: expectedNonceClaim,
	})
	if err != nil {
		return
	}

	// (Audience, Subject) is the verified Apple identity. Matching it to an
	// application user and issuing an application session remain backend work.
	identityKey := [2]string{result.Identity.Audience, result.Identity.Subject}
	_ = identityKey
}
