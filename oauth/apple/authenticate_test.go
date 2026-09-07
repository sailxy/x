package apple

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/sailxy/x/oauth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthenticate(t *testing.T) {
	identityKey := testIdentityPrivateKey(t)
	identityToken := signTestIdentityToken(t, validIdentityClaims(), "apple-key", jwt.SigningMethodRS256, identityKey)
	client := newLocalAppleClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/auth/token":
			require.NoError(t, r.ParseForm())
			assert.Equal(t, "com.example.web", r.Form.Get("client_id"))
			assert.Equal(t, "authorization-code", r.Form.Get("code"))
			assert.Equal(t, "https://example.com/apple/callback", r.Form.Get("redirect_uri"))
			require.NoError(t, json.NewEncoder(w).Encode(map[string]any{
				"access_token":  "apple-access-token",
				"refresh_token": "apple-refresh-token",
				"id_token":      identityToken.Value(),
				"token_type":    "Bearer",
				"expires_in":    3600,
			}))
		case "/auth/keys":
			writeTestJWKS(t, w, identityKey, "apple-key")
		default:
			http.NotFound(w, r)
		}
	}))

	result, err := client.Authenticate(context.Background(), AuthenticateRequest{
		ClientID:           "com.example.web",
		Code:               "authorization-code",
		RedirectURI:        "https://example.com/apple/callback",
		ExpectedNonceClaim: "expected-nonce",
	})
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "apple-subject", result.Identity.Subject)
	assert.Equal(t, "com.example.web", result.Identity.Audience)
	assert.Equal(t, "relay@example.com", result.Identity.Email)
	assert.True(t, result.Identity.EmailVerified)
	assert.True(t, result.Identity.IsPrivateEmail)
	assert.Equal(t, "expected-nonce", result.Identity.Nonce)
	assert.Equal(t, "apple-access-token", result.Token.AccessToken.Value())
	assert.Equal(t, "apple-refresh-token", result.Token.RefreshToken.Value())
	assert.Equal(t, identityToken.Value(), result.Token.IdentityToken.Value())
	assert.Equal(t, "Bearer", result.Token.TokenType)
	assert.Equal(t, int64(3600), result.Token.ExpiresIn)

	formatted := fmt.Sprintf("%+v", result)
	encoded, err := json.Marshal(result)
	require.NoError(t, err)
	for _, secret := range []string{"apple-access-token", "apple-refresh-token", identityToken.Value()} {
		assert.NotContains(t, formatted, secret)
		assert.NotContains(t, string(encoded), secret)
	}
}

func TestAuthenticateBindsExchangeAndVerificationClientID(t *testing.T) {
	identityKey := testIdentityPrivateKey(t)
	claims := validIdentityClaims()
	claims["aud"] = "com.example.ios"
	identityToken := signTestIdentityToken(t, claims, "apple-key", jwt.SigningMethodRS256, identityKey)
	client := newLocalAppleClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/auth/token":
			require.NoError(t, r.ParseForm())
			assert.Equal(t, "com.example.web", r.Form.Get("client_id"))
			require.NoError(t, json.NewEncoder(w).Encode(map[string]string{
				"id_token": identityToken.Value(),
			}))
		case "/auth/keys":
			writeTestJWKS(t, w, identityKey, "apple-key")
		default:
			http.NotFound(w, r)
		}
	}))

	result, err := client.Authenticate(context.Background(), AuthenticateRequest{
		ClientID: "com.example.web",
		Code:     "authorization-code",
	})
	assert.Nil(t, result)
	assert.ErrorIs(t, err, oauth.ErrTokenValidation)
}
