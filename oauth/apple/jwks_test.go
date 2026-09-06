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
	"sync/atomic"
	"testing"
	"time"

	"github.com/sailxy/x/oauth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPublicKeyCacheAndRotation(t *testing.T) {
	firstKey := testIdentityPrivateKey(t)
	secondKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	var requests atomic.Int32
	client := newLocalAppleClient(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Cache-Control", "public, max-age=3600")
		if requests.Add(1) == 1 {
			writeTestJWKS(t, w, firstKey, "first-key")
			return
		}
		writeTestJWKS(t, w, secondKey, "second-key")
	}))

	key, err := client.publicKey(context.Background(), "first-key")
	require.NoError(t, err)
	assert.Equal(t, firstKey.N, key.N)
	key, err = client.publicKey(context.Background(), "first-key")
	require.NoError(t, err)
	assert.Equal(t, firstKey.N, key.N)
	assert.Equal(t, int32(1), requests.Load())

	key, err = client.publicKey(context.Background(), "second-key")
	require.NoError(t, err)
	assert.Equal(t, secondKey.N, key.N)
	key, err = client.publicKey(context.Background(), "second-key")
	require.NoError(t, err)
	assert.Equal(t, int32(2), requests.Load())

	key, err = client.publicKey(context.Background(), "unknown-key")
	assert.Nil(t, key)
	assert.ErrorIs(t, err, oauth.ErrTokenValidation)
	assert.Equal(t, int32(3), requests.Load())
}

func TestPublicKeyConcurrentRefresh(t *testing.T) {
	key := testIdentityPrivateKey(t)
	started := make(chan struct{})
	release := make(chan struct{})
	var startOnce sync.Once
	var requests atomic.Int32
	client := newLocalAppleClient(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		requests.Add(1)
		startOnce.Do(func() { close(started) })
		<-release
		writeTestJWKS(t, w, key, "apple-key")
	}))

	const callers = 20
	errs := make(chan error, callers)
	var wait sync.WaitGroup
	for range callers {
		wait.Add(1)
		go func() {
			defer wait.Done()
			_, err := client.publicKey(context.Background(), "apple-key")
			errs <- err
		}()
	}
	<-started
	close(release)
	wait.Wait()
	close(errs)
	for err := range errs {
		assert.NoError(t, err)
	}
	assert.Equal(t, int32(1), requests.Load())
}

func TestPublicKeyWaitRespectsCancellation(t *testing.T) {
	key := testIdentityPrivateKey(t)
	started := make(chan struct{})
	release := make(chan struct{})
	client := newLocalAppleClient(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		close(started)
		<-release
		writeTestJWKS(t, w, key, "apple-key")
	}))

	firstDone := make(chan error, 1)
	go func() {
		_, err := client.publicKey(context.Background(), "apple-key")
		firstDone <- err
	}()
	<-started
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	publicKey, err := client.publicKey(ctx, "apple-key")
	assert.Nil(t, publicKey)
	assert.ErrorIs(t, err, oauth.ErrTransport)
	assert.ErrorIs(t, err, context.Canceled)
	close(release)
	assert.NoError(t, <-firstDone)
}

func TestPublicKeyRejectsInvalidResponses(t *testing.T) {
	validJWK := testJWK(testIdentityPrivateKey(t), "apple-key")
	invalidExponent := validJWK
	invalidExponent.Exponent = base64.RawURLEncoding.EncodeToString(big.NewInt(2).Bytes())
	duplicate, err := json.Marshal(jwkSet{Keys: []jwk{validJWK, validJWK}})
	require.NoError(t, err)
	invalidMetadata := validJWK
	invalidMetadata.KeyType = "EC"
	invalidModulus := validJWK
	invalidModulus.Modulus = "!"

	tests := []struct {
		name   string
		status int
		body   string
		target error
	}{
		{name: "non-success status", status: http.StatusBadGateway, body: `{}`, target: oauth.ErrPlatform},
		{name: "malformed JSON", status: http.StatusOK, body: `{`, target: oauth.ErrDecode},
		{name: "empty key set", status: http.StatusOK, body: `{"keys":[]}`, target: oauth.ErrInvalidResponse},
		{name: "unsupported key", status: http.StatusOK, body: marshalTestJSON(t, jwkSet{Keys: []jwk{invalidMetadata}}), target: oauth.ErrInvalidResponse},
		{name: "invalid modulus", status: http.StatusOK, body: marshalTestJSON(t, jwkSet{Keys: []jwk{invalidModulus}}), target: oauth.ErrInvalidResponse},
		{name: "invalid exponent", status: http.StatusOK, body: marshalTestJSON(t, jwkSet{Keys: []jwk{invalidExponent}}), target: oauth.ErrInvalidResponse},
		{name: "duplicate key ID", status: http.StatusOK, body: string(duplicate), target: oauth.ErrInvalidResponse},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			client := newLocalAppleClient(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(test.status)
				_, _ = io.WriteString(w, test.body)
			}))
			key, err := client.publicKey(context.Background(), "apple-key")
			assert.Nil(t, key)
			assert.ErrorIs(t, err, test.target)
		})
	}
}

func TestFailedRefreshDoesNotReplaceCache(t *testing.T) {
	key := testIdentityPrivateKey(t)
	var requests atomic.Int32
	client := newLocalAppleClient(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if requests.Add(1) == 1 {
			w.Header().Set("Cache-Control", "max-age=0")
			writeTestJWKS(t, w, key, "apple-key")
			return
		}
		_, _ = io.WriteString(w, `{"keys":[]}`)
	}))

	publicKey, err := client.publicKey(context.Background(), "apple-key")
	require.NoError(t, err)
	client.keys.mu.Lock()
	cached := client.keys.keys["apple-key"]
	client.keys.mu.Unlock()
	assert.Same(t, publicKey, cached)

	publicKey, err = client.publicKey(context.Background(), "apple-key")
	assert.Nil(t, publicKey)
	assert.ErrorIs(t, err, oauth.ErrInvalidResponse)
	client.keys.mu.Lock()
	assert.Same(t, cached, client.keys.keys["apple-key"])
	client.keys.mu.Unlock()
}

func TestKeysLifetime(t *testing.T) {
	for _, test := range []struct {
		name     string
		value    string
		expected time.Duration
	}{
		{name: "missing", expected: defaultKeysLifetime},
		{name: "max age", value: "public, max-age=300", expected: 5 * time.Minute},
		{name: "quoted", value: `max-age="60"`, expected: time.Minute},
		{name: "zero", value: "max-age=0", expected: 0},
		{name: "capped", value: "max-age=999999", expected: maxKeysLifetime},
		{name: "invalid", value: "max-age=invalid", expected: defaultKeysLifetime},
	} {
		t.Run(test.name, func(t *testing.T) {
			header := http.Header{}
			if test.value != "" {
				header.Set("Cache-Control", test.value)
			}
			assert.Equal(t, test.expected, keysLifetime(header))
		})
	}
}

func marshalTestJSON(t *testing.T, value any) string {
	t.Helper()
	payload, err := json.Marshal(value)
	require.NoError(t, err)
	return string(payload)
}
