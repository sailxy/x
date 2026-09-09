package apple

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/sailxy/x/oauth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthorizationURL(t *testing.T) {
	client := testAppleClient(t)
	address, err := client.AuthorizationURL(AuthorizationRequest{
		ClientID:    "com.example.web",
		RedirectURI: "https://example.com/callback?from=apple&next=/account",
		State:       "state + / ? & =",
		Nonce:       "nonce + / ? & =",
		Scopes:      []string{"name", "email", "name"},
	})
	require.NoError(t, err)

	parsed, err := url.Parse(address)
	require.NoError(t, err)
	assert.Equal(t, "https", parsed.Scheme)
	assert.Equal(t, "appleid.apple.com", parsed.Host)
	assert.Equal(t, "/auth/authorize", parsed.Path)
	assert.Equal(t, url.Values{
		"client_id":     {"com.example.web"},
		"redirect_uri":  {"https://example.com/callback?from=apple&next=/account"},
		"response_type": {"code"},
		"response_mode": {"form_post"},
		"state":         {"state + / ? & ="},
		"nonce":         {"nonce + / ? & ="},
		"scope":         {"name email"},
	}, parsed.Query())
}

func TestAuthorizationURLWithoutOptionalValues(t *testing.T) {
	client := testAppleClient(t)
	address, err := client.AuthorizationURL(AuthorizationRequest{
		ClientID:    "com.example.web",
		RedirectURI: "https://example.com/callback",
		State:       "state",
	})
	require.NoError(t, err)

	parsed, err := url.Parse(address)
	require.NoError(t, err)
	assert.NotContains(t, parsed.Query(), "nonce")
	assert.NotContains(t, parsed.Query(), "scope")
	assert.Equal(t, "form_post", parsed.Query().Get("response_mode"))
}

func TestAuthorizationURLRejectsInvalidInput(t *testing.T) {
	client := testAppleClient(t)
	valid := func() AuthorizationRequest {
		return AuthorizationRequest{
			ClientID:    "com.example.web",
			RedirectURI: "https://example.com/callback",
			State:       "state",
		}
	}
	tests := []struct {
		name   string
		mutate func(*AuthorizationRequest)
	}{
		{name: "missing client ID", mutate: func(r *AuthorizationRequest) { r.ClientID = "" }},
		{name: "unknown client ID", mutate: func(r *AuthorizationRequest) { r.ClientID = "com.example.unknown" }},
		{name: "missing redirect URI", mutate: func(r *AuthorizationRequest) { r.RedirectURI = "  " }},
		{name: "missing state", mutate: func(r *AuthorizationRequest) { r.State = "" }},
		{name: "unsupported scope", mutate: func(r *AuthorizationRequest) { r.Scopes = []string{"openid"} }},
		{name: "blank scope", mutate: func(r *AuthorizationRequest) { r.Scopes = []string{"email", " "} }},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := valid()
			test.mutate(&request)
			address, err := client.AuthorizationURL(request)
			assert.Empty(t, address)
			assert.ErrorIs(t, err, oauth.ErrInvalidInput)

			var oauthError *oauth.Error
			require.True(t, errors.As(err, &oauthError))
			assert.Equal(t, oauth.ProviderApple, oauthError.Provider)
			assert.Equal(t, "create authorization URL", oauthError.Operation)
		})
	}
}

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
			require.NoError(t, json.NewEncoder(w).Encode(map[string]string{"id_token": identityToken.Value()}))
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

func TestNew(t *testing.T) {
	privateKey := testECPrivateKey(t, elliptic.P256())
	privateKeyPEM := testPKCS8PEM(t, privateKey)
	httpClient := &http.Client{Timeout: time.Second}

	client, err := New(Config{
		TeamID:        " TEAM123456 ",
		ClientIDs:     []string{"com.example.web", "com.example.web", " com.example.ios "},
		KeyID:         " KEY1234567 ",
		PrivateKeyPEM: privateKeyPEM,
		HTTPClient:    httpClient,
	})
	require.NoError(t, err)
	require.NotNil(t, client)
	assert.Equal(t, "TEAM123456", client.teamID)
	assert.Equal(t, "KEY1234567", client.keyID)
	assert.Len(t, client.clientIDs, 2)
	assert.Contains(t, client.clientIDs, "com.example.web")
	assert.Contains(t, client.clientIDs, "com.example.ios")
	assert.Equal(t, privateKey.D, client.privateKey.D)
	assert.NotNil(t, client.rest)
	assert.WithinDuration(t, time.Now(), client.now(), time.Second)
}

func TestNewRejectsInvalidConfig(t *testing.T) {
	validKey := testPKCS8PEM(t, testECPrivateKey(t, elliptic.P256()))
	p384Key := testPKCS8PEM(t, testECPrivateKey(t, elliptic.P384()))
	rsaKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	rsaPEM := testPKCS8PEM(t, rsaKey)
	sec1Bytes, err := x509.MarshalECPrivateKey(testECPrivateKey(t, elliptic.P256()))
	require.NoError(t, err)
	sec1PEM := oauth.SensitiveString(pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: sec1Bytes}))

	valid := func() Config {
		return Config{
			TeamID:        "TEAM123456",
			ClientIDs:     []string{"com.example.web"},
			KeyID:         "KEY1234567",
			PrivateKeyPEM: validKey,
		}
	}
	tests := []struct {
		name   string
		mutate func(*Config)
	}{
		{name: "missing team ID", mutate: func(c *Config) { c.TeamID = "" }},
		{name: "blank team ID", mutate: func(c *Config) { c.TeamID = "  " }},
		{name: "missing client IDs", mutate: func(c *Config) { c.ClientIDs = nil }},
		{name: "blank client ID", mutate: func(c *Config) { c.ClientIDs = []string{"com.example.web", " "} }},
		{name: "missing key ID", mutate: func(c *Config) { c.KeyID = "" }},
		{name: "blank key ID", mutate: func(c *Config) { c.KeyID = "  " }},
		{name: "missing private key", mutate: func(c *Config) { c.PrivateKeyPEM = "" }},
		{name: "invalid PEM", mutate: func(c *Config) { c.PrivateKeyPEM = "not-a-private-key" }},
		{name: "wrong PEM type", mutate: func(c *Config) { c.PrivateKeyPEM = sec1PEM }},
		{name: "invalid PKCS8", mutate: func(c *Config) {
			c.PrivateKeyPEM = oauth.SensitiveString(pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: []byte("invalid")}))
		}},
		{name: "RSA private key", mutate: func(c *Config) { c.PrivateKeyPEM = rsaPEM }},
		{name: "P384 private key", mutate: func(c *Config) { c.PrivateKeyPEM = p384Key }},
		{name: "trailing PEM data", mutate: func(c *Config) { c.PrivateKeyPEM = oauth.SensitiveString(validKey.Value() + "trailing") }},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			config := valid()
			test.mutate(&config)
			client, err := New(config)
			assert.Nil(t, client)
			assert.ErrorIs(t, err, oauth.ErrInvalidConfig)

			var oauthError *oauth.Error
			require.True(t, errors.As(err, &oauthError))
			assert.Equal(t, oauth.ProviderApple, oauthError.Provider)
			assert.Equal(t, "create client", oauthError.Operation)
			if config.PrivateKeyPEM.Value() != "" {
				assert.NotContains(t, err.Error(), config.PrivateKeyPEM.Value())
			}
		})
	}
}

func TestFormattingOmitsApplePrivateKey(t *testing.T) {
	privateKey := testECPrivateKey(t, elliptic.P256())
	privateKeyPEM := testPKCS8PEM(t, privateKey)
	config := Config{
		TeamID:        "TEAM123456",
		ClientIDs:     []string{"com.example.web"},
		KeyID:         "KEY1234567",
		PrivateKeyPEM: privateKeyPEM,
	}
	client, err := New(config)
	require.NoError(t, err)

	for _, value := range []any{config, client, *client} {
		for _, format := range []string{"%s", "%q", "%v", "%+v", "%#v"} {
			formatted := fmt.Sprintf(format, value)
			assert.NotContains(t, formatted, privateKeyPEM.Value())
			assert.NotContains(t, formatted, privateKey.D.Text(16))
		}
	}
}

func testECPrivateKey(t *testing.T, curve elliptic.Curve) *ecdsa.PrivateKey {
	t.Helper()
	privateKey, err := ecdsa.GenerateKey(curve, rand.Reader)
	require.NoError(t, err)
	return privateKey
}

func testPKCS8PEM(t *testing.T, privateKey any) oauth.SensitiveString {
	t.Helper()
	der, err := x509.MarshalPKCS8PrivateKey(privateKey)
	require.NoError(t, err)
	return oauth.SensitiveString(pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der}))
}

func testAppleClient(t *testing.T) *Client {
	t.Helper()
	client, err := New(Config{
		TeamID:        "TEAM123456",
		ClientIDs:     []string{"com.example.web", "com.example.ios"},
		KeyID:         "KEY1234567",
		PrivateKeyPEM: testPKCS8PEM(t, testECPrivateKey(t, elliptic.P256())),
	})
	require.NoError(t, err)
	client.now = testAppleNow
	return client
}
