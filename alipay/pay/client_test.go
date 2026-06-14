package pay

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	got, err := New(Config{
		AppID:              os.Getenv("ALIPAY_APP_ID"),
		PrivateKeyPEM:      os.Getenv("ALIPAY_PRIVATE_KEY_PEM"),
		AlipayPublicKeyPEM: os.Getenv("ALIPAY_PUBLIC_KEY_PEM"),
		NotifyURL:          os.Getenv("ALIPAY_NOTIFY_URL"),
		ReturnURL:          os.Getenv("ALIPAY_RETURN_URL"),
	})

	require.NoError(t, err)
	require.NotNil(t, got)

	resp, err := got.PagePay(context.Background(), PagePayRequest{
		PayRequest: PayRequest{
			Subject:     "test",
			OutTradeNo:  "1234567890",
			TotalAmount: "0.01",
		},
		ReturnURL: "https://example.com/return",
	})

	require.NoError(t, err)
	require.NotNil(t, resp)
	t.Logf("resp: %+v", resp)
}

func TestNewRequiresOrdinaryMerchantCredentials(t *testing.T) {
	tests := []struct {
		name string
		cfg  Config
	}{
		{name: "missing app id", cfg: validConfig(func(c *Config) { c.AppID = "" })},
		{name: "missing private key source", cfg: validConfig(func(c *Config) { c.PrivateKeyPEM = "" })},
		{name: "missing alipay public key", cfg: validConfig(func(c *Config) { c.AlipayPublicKeyPEM = "" })},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := newWithFactory(tt.cfg, noopFactory)

			assert.Nil(t, got)
			assert.Error(t, err)
		})
	}
}

func TestNewLoadsPrivateKeyFromPEM(t *testing.T) {
	got, err := newWithFactory(validConfig(), noopFactory)

	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, "app-123", got.appID)
	assert.False(t, got.production)
}

func TestNewLoadsPrivateKeyFromPath(t *testing.T) {
	keyPath := writeTempFile(t, testPrivateKeyPEM(t))
	cfg := validConfig(func(c *Config) {
		c.PrivateKeyPEM = ""
		c.PrivateKeyPath = keyPath
		c.Sandbox = false
	})

	got, err := newWithFactory(cfg, noopFactory)

	require.NoError(t, err)
	require.NotNil(t, got)
	assert.True(t, got.production)
}

func TestNewLoadsBareBase64PrivateKeyFromPath(t *testing.T) {
	keyPath := writeTempFile(t, testPrivateKeyBase64(t))
	cfg := validConfig(func(c *Config) {
		c.PrivateKeyPEM = ""
		c.PrivateKeyPath = keyPath
	})

	got, err := newWithFactory(cfg, noopFactory)

	require.NoError(t, err)
	require.NotNil(t, got)
}

func TestNewRejectsInvalidPrivateKey(t *testing.T) {
	cfg := validConfig(func(c *Config) {
		c.PrivateKeyPEM = "not a private key"
	})

	got, err := newWithFactory(cfg, noopFactory)

	assert.Nil(t, got)
	assert.Error(t, err)
}

func validConfig(mutators ...func(*Config)) Config {
	privateKey := testPrivateKeyPEM(nil)
	publicKey := testPublicKeyPEM(nil)
	cfg := Config{
		AppID:              "app-123",
		PrivateKeyPEM:      privateKey,
		AlipayPublicKeyPEM: publicKey,
		NotifyURL:          "https://example.com/alipay/notify",
		ReturnURL:          "https://example.com/pay/return",
		Sandbox:            true,
	}
	for _, mutate := range mutators {
		mutate(&cfg)
	}
	return cfg
}

func testPrivateKeyPEM(t *testing.T) string {
	key := testRSAKey(t)
	der, err := x509.MarshalPKCS8PrivateKey(key)
	if t != nil {
		require.NoError(t, err)
	} else if err != nil {
		panic(err)
	}
	return string(pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der}))
}

func testPrivateKeyBase64(t *testing.T) string {
	key := testRSAKey(t)
	der, err := x509.MarshalPKCS8PrivateKey(key)
	if t != nil {
		require.NoError(t, err)
	} else if err != nil {
		panic(err)
	}
	return base64.StdEncoding.EncodeToString(der)
}

func testPublicKeyPEM(t *testing.T) string {
	key := testRSAKey(t)
	der, err := x509.MarshalPKIXPublicKey(&key.PublicKey)
	if t != nil {
		require.NoError(t, err)
	} else if err != nil {
		panic(err)
	}
	return string(pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: der}))
}

func testRSAKey(t *testing.T) *rsa.PrivateKey {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if t != nil {
		require.NoError(t, err)
	} else if err != nil {
		panic(err)
	}
	return key
}

func writeTempFile(t *testing.T, content string) string {
	t.Helper()

	file, err := os.CreateTemp(t.TempDir(), "alipay-key-*.pem")
	require.NoError(t, err)
	_, err = file.WriteString(content)
	require.NoError(t, err)
	require.NoError(t, file.Close())
	return file.Name()
}

func noopFactory(Config, string) (*clientParts, error) {
	return &clientParts{}, nil
}
