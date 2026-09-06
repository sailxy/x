package apple

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClientSecret(t *testing.T) {
	privateKey := fixedTestPrivateKey()
	now := time.Date(2026, time.September, 7, 12, 0, 0, 0, time.UTC)
	client, err := New(Config{
		TeamID:        "TEAM123456",
		ClientIDs:     []string{"com.example.web"},
		KeyID:         "KEY1234567",
		PrivateKeyPEM: testPKCS8PEM(t, privateKey),
	})
	require.NoError(t, err)
	client.now = func() time.Time { return now }

	secret, err := client.clientSecret("com.example.web")
	require.NoError(t, err)
	assert.NotEmpty(t, secret.Value())
	assert.NotContains(t, fmt.Sprintf("%v", secret), secret.Value())

	claims := &jwt.RegisteredClaims{}
	token, err := jwt.ParseWithClaims(secret.Value(), claims, func(token *jwt.Token) (any, error) {
		assert.Same(t, jwt.SigningMethodES256, token.Method)
		return &privateKey.PublicKey, nil
	},
		jwt.WithValidMethods([]string{jwt.SigningMethodES256.Alg()}),
		jwt.WithIssuer("TEAM123456"),
		jwt.WithAudience(appleIssuer),
		jwt.WithTimeFunc(func() time.Time { return now }),
	)
	require.NoError(t, err)
	assert.True(t, token.Valid)
	assert.Equal(t, "KEY1234567", token.Header["kid"])
	assert.Equal(t, "TEAM123456", claims.Issuer)
	assert.Equal(t, "com.example.web", claims.Subject)
	assert.Equal(t, jwt.ClaimStrings{appleIssuer}, claims.Audience)
	assert.True(t, claims.IssuedAt.Time.Equal(now))
	assert.True(t, claims.ExpiresAt.Time.Equal(now.Add(clientSecretLifetime)))
}

// fixedTestPrivateKey is deterministic test data, not a production credential.
func fixedTestPrivateKey() *ecdsa.PrivateKey {
	d := big.NewInt(1)
	x, y := elliptic.P256().ScalarBaseMult(d.Bytes())
	return &ecdsa.PrivateKey{
		PublicKey: ecdsa.PublicKey{Curve: elliptic.P256(), X: x, Y: y},
		D:         d,
	}
}
