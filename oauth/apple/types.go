package apple

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

// AuthenticateResult contains a verified Apple identity and non-sensitive
// metadata from the token exchange.
type AuthenticateResult struct {
	Identity  Identity
	TokenType string
	ExpiresIn int64
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
