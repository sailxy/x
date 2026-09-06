package apple

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"math/big"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/sailxy/x/oauth"
)

const (
	keysEndpoint        = "https://appleid.apple.com/auth/keys"
	defaultKeysLifetime = time.Hour
	maxKeysLifetime     = 24 * time.Hour
)

type jwksCache struct {
	mu         sync.Mutex
	keys       map[string]*rsa.PublicKey
	expiresAt  time.Time
	refreshing bool
	refreshed  chan struct{}
	refreshErr error
}

type jwkSet struct {
	Keys []jwk `json:"keys"`
}

type jwk struct {
	KeyType   string `json:"kty"`
	KeyID     string `json:"kid"`
	Use       string `json:"use"`
	Algorithm string `json:"alg"`
	Modulus   string `json:"n"`
	Exponent  string `json:"e"`
}

func (c *Client) publicKey(ctx context.Context, keyID string) (*rsa.PublicKey, error) {
	if ctx == nil {
		return nil, invalidInput("fetch public keys", errors.New("context is required"))
	}
	if strings.TrimSpace(keyID) == "" {
		return nil, tokenValidation(errors.New("key ID is required"))
	}

	c.keys.mu.Lock()
	if key := c.keys.keys[keyID]; key != nil && c.now().Before(c.keys.expiresAt) {
		c.keys.mu.Unlock()
		return key, nil
	}
	if c.keys.refreshing {
		refreshed := c.keys.refreshed
		c.keys.mu.Unlock()
		select {
		case <-ctx.Done():
			return nil, oauth.NewError(oauth.ProviderApple, oauth.ErrorKindTransport, "fetch public keys", ctx.Err())
		case <-refreshed:
		}
		c.keys.mu.Lock()
		defer c.keys.mu.Unlock()
		if c.keys.refreshErr != nil {
			return nil, c.keys.refreshErr
		}
		if key := c.keys.keys[keyID]; key != nil {
			return key, nil
		}
		return nil, tokenValidation(errors.New("public key is unavailable"))
	}

	c.keys.refreshing = true
	c.keys.refreshed = make(chan struct{})
	refreshed := c.keys.refreshed
	c.keys.mu.Unlock()

	keys, expiresAt, err := c.fetchPublicKeys(ctx)
	c.keys.mu.Lock()
	if err == nil {
		c.keys.keys = keys
		c.keys.expiresAt = expiresAt
	}
	c.keys.refreshErr = err
	c.keys.refreshing = false
	close(refreshed)
	key := c.keys.keys[keyID]
	c.keys.mu.Unlock()

	if err != nil {
		return nil, err
	}
	if key == nil {
		return nil, tokenValidation(errors.New("public key is unavailable"))
	}
	return key, nil
}

func (c *Client) fetchPublicKeys(ctx context.Context) (map[string]*rsa.PublicKey, time.Time, error) {
	req, err := http.NewRequest(http.MethodGet, keysEndpoint, nil)
	if err != nil {
		return nil, time.Time{}, oauth.NewError(
			oauth.ProviderApple,
			oauth.ErrorKindInvalidConfig,
			"fetch public keys",
			err,
		)
	}
	req.Header.Set("Accept", "application/json")

	response, requestErr := c.do(ctx, "fetch public keys", req)
	if requestErr != nil && response == nil {
		return nil, time.Time{}, requestErr
	}
	if requestErr != nil {
		return nil, time.Time{}, requestErr
	}

	var payload jwkSet
	if err := json.Unmarshal(response.Body, &payload); err != nil {
		return nil, time.Time{}, oauth.NewError(
			oauth.ProviderApple,
			oauth.ErrorKindDecode,
			"fetch public keys",
			err,
		)
	}
	keys, err := parsePublicKeys(payload)
	if err != nil {
		return nil, time.Time{}, err
	}
	return keys, c.now().Add(keysLifetime(response.Header)), nil
}

func parsePublicKeys(set jwkSet) (map[string]*rsa.PublicKey, error) {
	if len(set.Keys) == 0 {
		return nil, invalidKeys(errors.New("public key set is empty"))
	}
	keys := make(map[string]*rsa.PublicKey, len(set.Keys))
	for _, value := range set.Keys {
		if value.KeyType != "RSA" || value.Use != "sig" || value.Algorithm != "RS256" || strings.TrimSpace(value.KeyID) == "" {
			return nil, invalidKeys(errors.New("public key metadata is invalid"))
		}
		if _, exists := keys[value.KeyID]; exists {
			return nil, invalidKeys(errors.New("public key ID is duplicated"))
		}
		modulus, err := base64.RawURLEncoding.DecodeString(value.Modulus)
		if err != nil {
			return nil, invalidKeys(errors.New("public key modulus is invalid"))
		}
		exponent, err := base64.RawURLEncoding.DecodeString(value.Exponent)
		if err != nil {
			return nil, invalidKeys(errors.New("public key exponent is invalid"))
		}
		n := new(big.Int).SetBytes(modulus)
		e := new(big.Int).SetBytes(exponent)
		if n.Sign() <= 0 || n.BitLen() < 2048 || !e.IsInt64() || e.Int64() < 3 || e.Bit(0) == 0 || e.Int64() > int64(^uint(0)>>1) {
			return nil, invalidKeys(errors.New("public key parameters are invalid"))
		}
		keys[value.KeyID] = &rsa.PublicKey{N: n, E: int(e.Int64())}
	}
	return keys, nil
}

func keysLifetime(header http.Header) time.Duration {
	for _, directive := range strings.Split(header.Get("Cache-Control"), ",") {
		name, value, ok := strings.Cut(strings.TrimSpace(directive), "=")
		if !ok || !strings.EqualFold(name, "max-age") {
			continue
		}
		seconds, err := strconv.ParseInt(strings.Trim(value, `"`), 10, 64)
		if err != nil || seconds < 0 {
			break
		}
		if seconds > int64(maxKeysLifetime/time.Second) {
			return maxKeysLifetime
		}
		return time.Duration(seconds) * time.Second
	}
	return defaultKeysLifetime
}

func invalidKeys(cause error) error {
	return oauth.NewError(oauth.ProviderApple, oauth.ErrorKindInvalidResponse, "fetch public keys", cause)
}

func tokenValidation(cause error) error {
	return oauth.NewError(oauth.ProviderApple, oauth.ErrorKindTokenValidation, "verify identity token", cause)
}
