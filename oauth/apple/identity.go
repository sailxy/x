package apple

import (
	"context"
	"crypto/subtle"
	"errors"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/sailxy/x/oauth"
)

const identityTokenLeeway = time.Minute

type identityClaims struct {
	jwt.RegisteredClaims
	Email          string    `json:"email"`
	EmailVerified  appleBool `json:"email_verified"`
	IsPrivateEmail appleBool `json:"is_private_email"`
	Nonce          string    `json:"nonce"`
}

type appleBool struct {
	value bool
}

func (b *appleBool) UnmarshalJSON(data []byte) error {
	switch strings.TrimSpace(string(data)) {
	case "true", `"true"`:
		b.value = true
		return nil
	case "false", `"false"`:
		b.value = false
		return nil
	}
	return errors.New("Apple boolean claim is invalid")
}

func (c *Client) verifyIdentityToken(
	ctx context.Context,
	clientID string,
	identityToken oauth.SensitiveString,
	expectedNonceClaim string,
) (*Identity, error) {
	claims := &identityClaims{}
	var keyErr error
	token, err := jwt.ParseWithClaims(
		identityToken.Value(),
		claims,
		func(token *jwt.Token) (any, error) {
			keyID, ok := token.Header["kid"].(string)
			if !ok || strings.TrimSpace(keyID) == "" {
				return nil, errors.New("identity token key ID is missing")
			}
			key, err := c.publicKey(ctx, keyID)
			if err != nil {
				keyErr = err
				return nil, err
			}
			return key, nil
		},
		jwt.WithValidMethods([]string{jwt.SigningMethodRS256.Alg()}),
		jwt.WithIssuer(appleIssuer),
		jwt.WithAudience(clientID),
		jwt.WithExpirationRequired(),
		jwt.WithIssuedAt(),
		jwt.WithLeeway(identityTokenLeeway),
		jwt.WithTimeFunc(c.now),
	)
	if keyErr != nil {
		return nil, keyErr
	}
	if err != nil || token == nil || !token.Valid {
		return nil, tokenValidation(err)
	}
	if strings.TrimSpace(claims.Subject) == "" {
		return nil, tokenValidation(errors.New("subject is required"))
	}
	if expectedNonceClaim != "" && subtle.ConstantTimeCompare(
		[]byte(claims.Nonce),
		[]byte(expectedNonceClaim),
	) != 1 {
		return nil, tokenValidation(errors.New("nonce does not match"))
	}

	return &Identity{
		Subject:        claims.Subject,
		Audience:       clientID,
		Email:          claims.Email,
		EmailVerified:  claims.EmailVerified.value,
		IsPrivateEmail: claims.IsPrivateEmail.value,
		Nonce:          claims.Nonce,
	}, nil
}
