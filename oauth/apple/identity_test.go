package apple

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"io"
	"math/big"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/sailxy/x/oauth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVerifyIdentityToken(t *testing.T) {
	key := testIdentityPrivateKey(t)
	client := newLocalAppleClient(t, testJWKSHandler(t, key, "apple-key"))
	claims := validIdentityClaims()
	claims["aud"] = []string{"another-client", "com.example.web"}
	token := signTestIdentityToken(t, claims, "apple-key", jwt.SigningMethodRS256, key)

	identity, err := client.verifyIdentityToken(context.Background(), "com.example.web", token, "expected-nonce")
	require.NoError(t, err)
	require.NotNil(t, identity)
	assert.Equal(t, "apple-subject", identity.Subject)
	assert.Equal(t, "com.example.web", identity.Audience)
	assert.Equal(t, "relay@example.com", identity.Email)
	assert.True(t, identity.EmailVerified)
	assert.True(t, identity.IsPrivateEmail)
	assert.Equal(t, "expected-nonce", identity.Nonce)
}

func TestVerifyIdentityTokenRejectsInvalidToken(t *testing.T) {
	trustedKey := testIdentityPrivateKey(t)
	untrustedKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	tests := []struct {
		name  string
		token func(*testing.T) oauth.SensitiveString
	}{
		{name: "wrong algorithm", token: func(t *testing.T) oauth.SensitiveString {
			return signTestIdentityToken(t, validIdentityClaims(), "apple-key", jwt.SigningMethodHS256, []byte("test-signing-key"))
		}},
		{name: "unsigned", token: func(t *testing.T) oauth.SensitiveString {
			return signTestIdentityToken(t, validIdentityClaims(), "apple-key", jwt.SigningMethodNone, jwt.UnsafeAllowNoneSignatureType)
		}},
		{name: "wrong signature", token: func(t *testing.T) oauth.SensitiveString {
			return signTestIdentityToken(t, validIdentityClaims(), "apple-key", jwt.SigningMethodRS256, untrustedKey)
		}},
		{name: "wrong issuer", token: func(t *testing.T) oauth.SensitiveString {
			claims := validIdentityClaims()
			claims["iss"] = "https://example.com"
			return signTestIdentityToken(t, claims, "apple-key", jwt.SigningMethodRS256, trustedKey)
		}},
		{name: "wrong audience", token: func(t *testing.T) oauth.SensitiveString {
			claims := validIdentityClaims()
			claims["aud"] = "com.example.ios"
			return signTestIdentityToken(t, claims, "apple-key", jwt.SigningMethodRS256, trustedKey)
		}},
		{name: "expired", token: func(t *testing.T) oauth.SensitiveString {
			claims := validIdentityClaims()
			claims["exp"] = testAppleNow().Add(-2 * identityTokenLeeway).Unix()
			return signTestIdentityToken(t, claims, "apple-key", jwt.SigningMethodRS256, trustedKey)
		}},
		{name: "missing expiration", token: func(t *testing.T) oauth.SensitiveString {
			claims := validIdentityClaims()
			delete(claims, "exp")
			return signTestIdentityToken(t, claims, "apple-key", jwt.SigningMethodRS256, trustedKey)
		}},
		{name: "future issued at", token: func(t *testing.T) oauth.SensitiveString {
			claims := validIdentityClaims()
			claims["iat"] = testAppleNow().Add(2 * identityTokenLeeway).Unix()
			return signTestIdentityToken(t, claims, "apple-key", jwt.SigningMethodRS256, trustedKey)
		}},
		{name: "missing subject", token: func(t *testing.T) oauth.SensitiveString {
			claims := validIdentityClaims()
			delete(claims, "sub")
			return signTestIdentityToken(t, claims, "apple-key", jwt.SigningMethodRS256, trustedKey)
		}},
		{name: "blank subject", token: func(t *testing.T) oauth.SensitiveString {
			claims := validIdentityClaims()
			claims["sub"] = "  "
			return signTestIdentityToken(t, claims, "apple-key", jwt.SigningMethodRS256, trustedKey)
		}},
		{name: "missing key ID", token: func(t *testing.T) oauth.SensitiveString {
			return signTestIdentityToken(t, validIdentityClaims(), "", jwt.SigningMethodRS256, trustedKey)
		}},
		{name: "malformed", token: func(*testing.T) oauth.SensitiveString {
			return oauth.SensitiveString("not-a-jwt")
		}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			client := newLocalAppleClient(t, testJWKSHandler(t, trustedKey, "apple-key"))
			identity, err := client.verifyIdentityToken(context.Background(), "com.example.web", test.token(t), "")
			assert.Nil(t, identity)
			assert.ErrorIs(t, err, oauth.ErrTokenValidation)
		})
	}
}

func TestVerifyIdentityTokenNonce(t *testing.T) {
	key := testIdentityPrivateKey(t)
	client := newLocalAppleClient(t, testJWKSHandler(t, key, "apple-key"))

	t.Run("disabled", func(t *testing.T) {
		claims := validIdentityClaims()
		delete(claims, "nonce")
		identity, err := client.verifyIdentityToken(
			context.Background(),
			"com.example.web",
			signTestIdentityToken(t, claims, "apple-key", jwt.SigningMethodRS256, key),
			"",
		)
		require.NoError(t, err)
		assert.Empty(t, identity.Nonce)
	})

	for _, test := range []struct {
		name  string
		nonce any
	}{
		{name: "missing", nonce: nil},
		{name: "mismatch", nonce: "different-nonce"},
	} {
		t.Run(test.name, func(t *testing.T) {
			claims := validIdentityClaims()
			if test.nonce == nil {
				delete(claims, "nonce")
			} else {
				claims["nonce"] = test.nonce
			}
			identity, err := client.verifyIdentityToken(
				context.Background(),
				"com.example.web",
				signTestIdentityToken(t, claims, "apple-key", jwt.SigningMethodRS256, key),
				"expected-nonce",
			)
			assert.Nil(t, identity)
			assert.ErrorIs(t, err, oauth.ErrTokenValidation)
		})
	}
}

func TestVerifyIdentityTokenAppleBooleanClaims(t *testing.T) {
	key := testIdentityPrivateKey(t)
	client := newLocalAppleClient(t, testJWKSHandler(t, key, "apple-key"))

	for _, test := range []struct {
		name             string
		emailVerified    any
		privateEmail     any
		expectedVerified bool
		expectedPrivate  bool
	}{
		{name: "booleans", emailVerified: true, privateEmail: false, expectedVerified: true},
		{name: "boolean strings", emailVerified: "false", privateEmail: "true", expectedPrivate: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			claims := validIdentityClaims()
			claims["email_verified"] = test.emailVerified
			claims["is_private_email"] = test.privateEmail
			identity, err := client.verifyIdentityToken(
				context.Background(),
				"com.example.web",
				signTestIdentityToken(t, claims, "apple-key", jwt.SigningMethodRS256, key),
				"",
			)
			require.NoError(t, err)
			assert.Equal(t, test.expectedVerified, identity.EmailVerified)
			assert.Equal(t, test.expectedPrivate, identity.IsPrivateEmail)
		})
	}

	for _, invalid := range []any{1, "TRUE", nil, map[string]any{"value": true}} {
		claims := validIdentityClaims()
		claims["email_verified"] = invalid
		identity, err := client.verifyIdentityToken(
			context.Background(),
			"com.example.web",
			signTestIdentityToken(t, claims, "apple-key", jwt.SigningMethodRS256, key),
			"",
		)
		assert.Nil(t, identity)
		assert.ErrorIs(t, err, oauth.ErrTokenValidation)
	}
}

func validIdentityClaims() jwt.MapClaims {
	return jwt.MapClaims{
		"iss":              appleIssuer,
		"aud":              "com.example.web",
		"exp":              testAppleNow().Add(time.Hour).Unix(),
		"iat":              testAppleNow().Unix(),
		"sub":              "apple-subject",
		"email":            "relay@example.com",
		"email_verified":   "true",
		"is_private_email": true,
		"nonce":            "expected-nonce",
	}
}

func signTestIdentityToken(t *testing.T, claims jwt.MapClaims, keyID string, method jwt.SigningMethod, key any) oauth.SensitiveString {
	t.Helper()
	token := jwt.NewWithClaims(method, claims)
	if keyID != "" {
		token.Header["kid"] = keyID
	}
	signed, err := token.SignedString(key)
	require.NoError(t, err)
	return oauth.SensitiveString(signed)
}

func testJWKSHandler(t *testing.T, key *rsa.PrivateKey, keyID string) http.Handler {
	t.Helper()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/auth/keys", r.URL.Path)
		writeTestJWKS(t, w, key, keyID)
	})
}

func writeTestJWKS(t *testing.T, w io.Writer, key *rsa.PrivateKey, keyID string) {
	t.Helper()
	payload, err := json.Marshal(jwkSet{Keys: []jwk{testJWK(key, keyID)}})
	require.NoError(t, err)
	_, err = w.Write(payload)
	require.NoError(t, err)
}

func testJWK(key *rsa.PrivateKey, keyID string) jwk {
	return jwk{
		KeyType:   "RSA",
		KeyID:     keyID,
		Use:       "sig",
		Algorithm: "RS256",
		Modulus:   base64.RawURLEncoding.EncodeToString(key.N.Bytes()),
		Exponent:  base64.RawURLEncoding.EncodeToString(big.NewInt(int64(key.E)).Bytes()),
	}
}

func testAppleNow() time.Time {
	return time.Date(2026, time.September, 7, 12, 0, 0, 0, time.UTC)
}

var (
	testIdentityKeyOnce sync.Once
	testIdentityKey     *rsa.PrivateKey
	testIdentityKeyErr  error
)

// testIdentityPrivateKey is generated once and used only as a local test key.
func testIdentityPrivateKey(t *testing.T) *rsa.PrivateKey {
	t.Helper()
	testIdentityKeyOnce.Do(func() {
		testIdentityKey, testIdentityKeyErr = rsa.GenerateKey(rand.Reader, 2048)
	})
	require.NoError(t, testIdentityKeyErr)
	return testIdentityKey
}
