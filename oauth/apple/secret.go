package apple

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/sailxy/x/oauth"
)

const (
	appleIssuer          = "https://appleid.apple.com"
	clientSecretLifetime = 5 * time.Minute
)

func (c *Client) clientSecret(clientID string) (oauth.SensitiveString, error) {
	now := c.now().UTC()
	token := jwt.NewWithClaims(jwt.SigningMethodES256, jwt.RegisteredClaims{
		Issuer:    c.teamID,
		Subject:   clientID,
		Audience:  jwt.ClaimStrings{appleIssuer},
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(clientSecretLifetime)),
	})
	token.Header["kid"] = c.keyID
	signed, err := token.SignedString(c.privateKey)
	if err != nil {
		return "", oauth.NewError(
			oauth.ProviderApple,
			oauth.ErrorKindInvalidConfig,
			"generate client secret",
			err,
		)
	}
	return oauth.SensitiveString(signed), nil
}
