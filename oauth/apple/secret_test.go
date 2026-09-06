package apple

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"errors"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/sailxy/x/oauth"
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
		Now:           func() time.Time { return now },
	})
	require.NoError(t, err)

	secret, err := client.clientSecret("com.example.web", 24*time.Hour)
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
	assert.True(t, claims.ExpiresAt.Time.Equal(now.Add(24*time.Hour)))
}

func TestClientSecretLifetimeBoundary(t *testing.T) {
	client := testAppleClient(t)

	secret, err := client.clientSecret("com.example.web", maxClientSecretLifetime)
	require.NoError(t, err)
	assert.NotEmpty(t, secret.Value())
}

func TestClientSecretRejectsInvalidInput(t *testing.T) {
	client := testAppleClient(t)
	tests := []struct {
		name     string
		clientID string
		lifetime time.Duration
	}{
		{name: "missing client ID", clientID: "", lifetime: time.Hour},
		{name: "unknown client ID", clientID: "com.example.unknown", lifetime: time.Hour},
		{name: "zero lifetime", clientID: "com.example.web", lifetime: 0},
		{name: "negative lifetime", clientID: "com.example.web", lifetime: -time.Second},
		{name: "over six months", clientID: "com.example.web", lifetime: maxClientSecretLifetime + time.Nanosecond},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			secret, err := client.clientSecret(test.clientID, test.lifetime)
			assert.Empty(t, secret.Value())
			assert.ErrorIs(t, err, oauth.ErrInvalidInput)

			var oauthError *oauth.Error
			require.True(t, errors.As(err, &oauthError))
			assert.Equal(t, oauth.ProviderApple, oauthError.Provider)
			assert.Equal(t, "generate client secret", oauthError.Operation)
		})
	}
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
