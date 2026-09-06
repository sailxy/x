package apple

import (
	"context"
	"crypto/subtle"
	"errors"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
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

// VerifyIdentityToken verifies an Apple identity token and returns only
// trusted claims. ClientID must be selected from the configured allowlist.
func (c *Client) VerifyIdentityToken(ctx context.Context, request VerifyIdentityTokenRequest) (*Identity, error) {
	if ctx == nil {
		return nil, invalidInput("verify identity token", errors.New("context is required"))
	}
	clientID, err := c.clientID(request.ClientID, "verify identity token")
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(request.IdentityToken.Value()) == "" {
		return nil, invalidInput("verify identity token", errors.New("identity token is required"))
	}

	claims := &identityClaims{}
	var keyErr error
	token, err := jwt.ParseWithClaims(
		request.IdentityToken.Value(),
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
	if request.ExpectedNonceClaim != "" && subtle.ConstantTimeCompare(
		[]byte(claims.Nonce),
		[]byte(request.ExpectedNonceClaim),
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
