package apple

import (
	"context"
	"crypto/rsa"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/MicahParks/jwkset"
	"github.com/sailxy/x/oauth"
)

const keysEndpoint = "https://appleid.apple.com/auth/keys"

type jwksCache struct {
	mu         sync.Mutex
	storage    jwkset.Storage
	expiresAt  time.Time
	refreshing bool
	refreshed  chan struct{}
	refreshErr error
}

func (c *Client) publicKey(ctx context.Context, keyID string) (*rsa.PublicKey, error) {
	if ctx == nil {
		return nil, invalidInput("fetch public keys", errors.New("context is required"))
	}
	if strings.TrimSpace(keyID) == "" {
		return nil, tokenValidation(errors.New("key ID is required"))
	}

	c.keys.mu.Lock()
	if c.keys.storage != nil && c.now().Before(c.keys.expiresAt) {
		storage := c.keys.storage
		c.keys.mu.Unlock()
		key, err := publicKeyFromStorage(ctx, storage, keyID)
		if err != nil || key != nil {
			return key, err
		}
		c.keys.mu.Lock()
	}
	if c.keys.refreshing {
		refreshed := c.keys.refreshed
		c.keys.mu.Unlock()
		select {
		case <-ctx.Done():
			return nil, newError(oauth.ErrorKindTransport, "fetch public keys", ctx.Err())
		case <-refreshed:
		}
		c.keys.mu.Lock()
		defer c.keys.mu.Unlock()
		if c.keys.refreshErr != nil {
			return nil, c.keys.refreshErr
		}
		return requiredPublicKey(ctx, c.keys.storage, keyID)
	}

	c.keys.refreshing = true
	c.keys.refreshed = make(chan struct{})
	refreshed := c.keys.refreshed
	c.keys.mu.Unlock()

	storage, expiresAt, err := c.fetchPublicKeys(ctx)
	c.keys.mu.Lock()
	if err == nil {
		c.keys.storage = storage
		c.keys.expiresAt = expiresAt
	}
	c.keys.refreshErr = err
	c.keys.refreshing = false
	close(refreshed)
	storage = c.keys.storage
	c.keys.mu.Unlock()

	if err != nil {
		return nil, err
	}
	return requiredPublicKey(ctx, storage, keyID)
}

func (c *Client) fetchPublicKeys(ctx context.Context) (jwkset.Storage, time.Time, error) {
	req, err := http.NewRequest(http.MethodGet, keysEndpoint, nil)
	if err != nil {
		return nil, time.Time{}, newError(oauth.ErrorKindInvalidConfig, "fetch public keys", err)
	}
	req.Header.Set("Accept", "application/json")

	response, requestErr := c.do(ctx, "fetch public keys", req)
	if requestErr != nil && response == nil {
		return nil, time.Time{}, requestErr
	}
	if requestErr != nil {
		return nil, time.Time{}, requestErr
	}

	var payload jwkset.JWKSMarshal
	if err := json.Unmarshal(response.Body, &payload); err != nil {
		return nil, time.Time{}, newError(oauth.ErrorKindDecode, "fetch public keys", err)
	}
	storage, err := payload.ToStorage()
	if err != nil {
		return nil, time.Time{}, invalidKeys(err)
	}
	if err := validatePublicKeys(storage); err != nil {
		return nil, time.Time{}, err
	}
	return storage, c.now().Add(time.Hour), nil
}

func validatePublicKeys(storage jwkset.Storage) error {
	keys, err := storage.KeyReadAll(context.Background())
	if err != nil {
		return invalidKeys(err)
	}
	if len(keys) == 0 {
		return invalidKeys(errors.New("public key set is empty"))
	}
	seen := make(map[string]struct{}, len(keys))
	for _, key := range keys {
		metadata := key.Marshal()
		if metadata.KTY != jwkset.KtyRSA || metadata.USE != jwkset.UseSig || metadata.ALG != jwkset.AlgRS256 || strings.TrimSpace(metadata.KID) == "" {
			return invalidKeys(errors.New("public key metadata is invalid"))
		}
		if _, exists := seen[metadata.KID]; exists {
			return invalidKeys(errors.New("public key ID is duplicated"))
		}
		if _, err := rsaPublicKey(key); err != nil {
			return err
		}
		seen[metadata.KID] = struct{}{}
	}
	return nil
}

func publicKeyFromStorage(ctx context.Context, storage jwkset.Storage, keyID string) (*rsa.PublicKey, error) {
	if storage == nil {
		return nil, tokenValidation(errors.New("public key is unavailable"))
	}
	key, err := storage.KeyRead(ctx, keyID)
	if errors.Is(err, jwkset.ErrKeyNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, invalidKeys(err)
	}
	return rsaPublicKey(key)
}

func requiredPublicKey(ctx context.Context, storage jwkset.Storage, keyID string) (*rsa.PublicKey, error) {
	key, err := publicKeyFromStorage(ctx, storage, keyID)
	if err != nil || key != nil {
		return key, err
	}
	return nil, tokenValidation(errors.New("public key is unavailable"))
}

func rsaPublicKey(key jwkset.JWK) (*rsa.PublicKey, error) {
	publicKey, ok := key.Key().(*rsa.PublicKey)
	if !ok || publicKey.N == nil || publicKey.N.BitLen() < 2048 || publicKey.E < 3 || publicKey.E%2 == 0 {
		return nil, invalidKeys(errors.New("public key parameters are invalid"))
	}
	return publicKey, nil
}

func invalidKeys(cause error) error {
	return newError(oauth.ErrorKindInvalidResponse, "fetch public keys", cause)
}

func tokenValidation(cause error) error {
	return newError(oauth.ErrorKindTokenValidation, "verify identity token", cause)
}
