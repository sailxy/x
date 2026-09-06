package apple

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/sailxy/x/oauth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
