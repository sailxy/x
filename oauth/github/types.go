package github

import "github.com/sailxy/x/oauth"

// AuthenticateResult contains all data obtained while authenticating a
// GitHub user. The caller decides which fields to use or persist.
type AuthenticateResult struct {
	Token Token
	User  User
}

// Token contains the result of a GitHub authorization-code exchange.
type Token struct {
	AccessToken oauth.SensitiveString
	Scope       string
	TokenType   string
}

// User contains the authenticated GitHub identity and available profile data.
type User struct {
	AccountID int64  `json:"id"`
	NodeID    string `json:"node_id"`
	Login     string `json:"login"`
	Name      string `json:"name"`
	AvatarURL string `json:"avatar_url"`
	Email     string `json:"email"`
}
