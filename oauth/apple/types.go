package apple

import "github.com/sailxy/x/oauth"

// AuthorizationRequest contains the caller-specific values used to create an
// Apple Web authorization URL. ClientID must be one of the configured IDs.
type AuthorizationRequest struct {
	ClientID    string
	RedirectURI string
	State       string
	Nonce       string
	Scopes      []string
}

// AuthenticateRequest contains the values needed to exchange an Apple
// authorization code and verify the returned identity token. An empty
// ExpectedNonceClaim disables nonce validation.
type AuthenticateRequest struct {
	ClientID           string
	Code               string
	RedirectURI        string
	ExpectedNonceClaim string
}

// AuthenticateResult contains the complete token exchange data and the
// verified Apple identity. The caller decides how to use or persist it.
type AuthenticateResult struct {
	Token    Token
	Identity Identity
}

// Token contains the result of an Apple authorization-code exchange.
// Sensitive fields are redacted during ordinary formatting and JSON encoding.
type Token struct {
	AccessToken   oauth.SensitiveString `json:"access_token"`
	RefreshToken  oauth.SensitiveString `json:"refresh_token"`
	IdentityToken oauth.SensitiveString `json:"id_token"`
	TokenType     string                `json:"token_type"`
	ExpiresIn     int64                 `json:"expires_in"`
}

// Identity contains claims from a verified Apple identity token.
type Identity struct {
	Subject        string
	Audience       string
	Email          string
	EmailVerified  bool
	IsPrivateEmail bool
	Nonce          string
}
