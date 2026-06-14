package pay

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	got, err := New(Config{
		AppID:                      os.Getenv("WEIXIN_PAY_APP_ID"),
		MchID:                      os.Getenv("WEIXIN_PAY_MCH_ID"),
		MchCertificateSerialNumber: os.Getenv("WEIXIN_PAY_CERTIFICATE_SERIAL_NUMBER"),
		MchPrivateKeyPath:          os.Getenv("WEIXIN_PAY_PRIVATE_KEY_PATH"),
		MchAPIv3Key:                os.Getenv("WEIXIN_PAY_API_V3_KEY"),
	})

	require.NoError(t, err)
	require.NotNil(t, got)

	resp, err := got.NativePrepay(context.Background(), PrepayRequest{
		Description: "test",
		OutTradeNo:  "111345243455353",
		NotifyURL:   "https://example.com/notify",
		Total:       1,
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
		{name: "missing mch id", cfg: validConfig(func(c *Config) { c.MchID = "" })},
		{name: "missing certificate serial", cfg: validConfig(func(c *Config) { c.MchCertificateSerialNumber = "" })},
		{name: "missing api v3 key", cfg: validConfig(func(c *Config) { c.MchAPIv3Key = "" })},
		{name: "missing private key source", cfg: validConfig(func(c *Config) { c.MchPrivateKeyPEM = "" })},
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
	assert.Equal(t, "wx-app", got.appID)
	assert.Equal(t, "mch-123", got.mchID)
}

func TestNewLoadsPrivateKeyFromPath(t *testing.T) {
	keyPath := writeTempPrivateKey(t, testPrivateKeyPEM(t))
	cfg := validConfig(func(c *Config) {
		c.MchPrivateKeyPEM = ""
		c.MchPrivateKeyPath = keyPath
	})

	got, err := newWithFactory(cfg, noopFactory)

	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, "wx-app", got.appID)
	assert.Equal(t, "mch-123", got.mchID)
}

func TestNewRejectsInvalidPrivateKey(t *testing.T) {
	cfg := validConfig(func(c *Config) {
		c.MchPrivateKeyPEM = "not a private key"
	})

	got, err := newWithFactory(cfg, noopFactory)

	assert.Nil(t, got)
	assert.Error(t, err)
}

func validConfig(mutators ...func(*Config)) Config {
	cfg := Config{
		AppID:                      "wx-app",
		MchID:                      "mch-123",
		MchCertificateSerialNumber: "serial-123",
		MchPrivateKeyPEM:           testPrivateKeyPEM(nil),
		MchAPIv3Key:                "12345678901234567890123456789012",
	}
	for _, mutate := range mutators {
		mutate(&cfg)
	}
	return cfg
}

func testPrivateKeyPEM(t *testing.T) string {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if t != nil {
		require.NoError(t, err)
	} else if err != nil {
		panic(err)
	}
	der, err := x509.MarshalPKCS8PrivateKey(key)
	if t != nil {
		require.NoError(t, err)
	} else if err != nil {
		panic(err)
	}
	return string(pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der}))
}

func writeTempPrivateKey(t *testing.T, content string) string {
	t.Helper()

	file, err := os.CreateTemp(t.TempDir(), "mch-key-*.pem")
	require.NoError(t, err)
	_, err = file.WriteString(content)
	require.NoError(t, err)
	require.NoError(t, file.Close())
	return file.Name()
}

func noopFactory(Config, *rsa.PrivateKey) (*clientParts, error) {
	return &clientParts{}, nil
}
