package apple

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/sailxy/x/oauth"
)

const (
	appleIssuer             = "https://appleid.apple.com"
	maxClientSecretLifetime = 180 * 24 * time.Hour
)

func (c *Client) clientSecret(clientID string, lifetime time.Duration) (oauth.SensitiveString, error) {
	clientID, err := c.clientID(clientID, "generate client secret")
	if err != nil {
		return "", err
	}
	if lifetime <= 0 || lifetime > maxClientSecretLifetime {
		return "", invalidInput("generate client secret", errors.New("lifetime must be positive and at most six months"))
	}

	now := c.now().UTC()
	token := jwt.NewWithClaims(jwt.SigningMethodES256, jwt.RegisteredClaims{
		Issuer:    c.teamID,
		Subject:   clientID,
		Audience:  jwt.ClaimStrings{appleIssuer},
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(lifetime)),
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
