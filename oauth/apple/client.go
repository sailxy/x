// Package apple implements Sign in with Apple protocol operations.
//
// It validates Apple credentials and returns platform identity data only. It
// does not manage application state, business users, or application sessions.
package apple

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/sailxy/x/oauth"
	"github.com/sailxy/x/rest"
)

// Config configures a Sign in with Apple client.
type Config struct {
	TeamID        string
	ClientIDs     []string
	KeyID         string
	PrivateKeyPEM oauth.SensitiveString
	HTTPClient    *http.Client
	Now           func() time.Time
}

// Format omits credentials and HTTP client internals from formatted output.
func (c Config) Format(state fmt.State, _ rune) {
	_, _ = fmt.Fprintf(state, "Sign in with Apple config for team %q", c.TeamID)
}

// Client performs Sign in with Apple protocol operations.
type Client struct {
	teamID     string
	clientIDs  map[string]struct{}
	keyID      string
	privateKey *ecdsa.PrivateKey
	rest       *rest.REST
	now        func() time.Time
	keys       *jwksCache
}

// New creates a Sign in with Apple client and validates its configuration.
func New(config Config) (*Client, error) {
	teamID, err := requiredIdentifier(config.TeamID, "team ID")
	if err != nil {
		return nil, err
	}
	keyID, err := requiredIdentifier(config.KeyID, "key ID")
	if err != nil {
		return nil, err
	}
	clientIDs, err := clientIDSet(config.ClientIDs)
	if err != nil {
		return nil, err
	}
	privateKey, err := parsePrivateKey(config.PrivateKeyPEM)
	if err != nil {
		return nil, err
	}

	httpClient := rest.NewREST()
	if config.HTTPClient != nil {
		httpClient = rest.NewRESTWithClient(config.HTTPClient)
	}
	now := config.Now
	if now == nil {
		now = time.Now
	}
	return &Client{
		teamID:     teamID,
		clientIDs:  clientIDs,
		keyID:      keyID,
		privateKey: privateKey,
		rest:       httpClient,
		now:        now,
		keys:       &jwksCache{},
	}, nil
}

// Format omits credentials and HTTP client internals from formatted output.
func (c Client) Format(state fmt.State, _ rune) {
	_, _ = fmt.Fprintf(state, "Sign in with Apple client for team %q", c.teamID)
}

func (c *Client) do(ctx context.Context, operation string, req *http.Request) (*rest.Response, error) {
	if ctx == nil {
		return nil, invalidInput(operation, errors.New("context is required"))
	}

	response, err := c.rest.Do(ctx, req)
	if err == nil {
		return response, nil
	}
	if errors.Is(err, rest.ErrResponseTooLarge) {
		return nil, oauth.NewError(
			oauth.ProviderApple,
			oauth.ErrorKindInvalidResponse,
			operation,
			err,
		)
	}

	var statusErr *rest.StatusError
	if errors.As(err, &statusErr) {
		platformErr := oauth.NewError(
			oauth.ProviderApple,
			oauth.ErrorKindPlatform,
			operation,
			err,
		)
		platformErr.StatusCode = statusErr.StatusCode
		return response, platformErr
	}
	return nil, oauth.NewError(
		oauth.ProviderApple,
		oauth.ErrorKindTransport,
		operation,
		err,
	)
}

func requiredIdentifier(value, name string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", oauth.NewError(
			oauth.ProviderApple,
			oauth.ErrorKindInvalidConfig,
			"create client",
			fmt.Errorf("%s is required", name),
		)
	}
	return value, nil
}

func clientIDSet(values []string) (map[string]struct{}, error) {
	clientIDs := make(map[string]struct{}, len(values))
	for _, value := range values {
		clientID, err := requiredIdentifier(value, "client ID")
		if err != nil {
			return nil, err
		}
		clientIDs[clientID] = struct{}{}
	}
	if len(clientIDs) == 0 {
		return nil, oauth.NewError(
			oauth.ProviderApple,
			oauth.ErrorKindInvalidConfig,
			"create client",
			errors.New("at least one client ID is required"),
		)
	}
	return clientIDs, nil
}

func parsePrivateKey(value oauth.SensitiveString) (*ecdsa.PrivateKey, error) {
	block, remaining := pem.Decode([]byte(value.Value()))
	if block == nil || block.Type != "PRIVATE KEY" || len(strings.TrimSpace(string(remaining))) != 0 {
		return nil, invalidPrivateKey(errors.New("private key must contain one PKCS#8 PEM block"))
	}
	parsed, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, invalidPrivateKey(err)
	}
	privateKey, ok := parsed.(*ecdsa.PrivateKey)
	if !ok || privateKey.Curve != elliptic.P256() {
		return nil, invalidPrivateKey(errors.New("private key must use ECDSA P-256"))
	}
	return privateKey, nil
}

func invalidPrivateKey(cause error) error {
	return oauth.NewError(
		oauth.ProviderApple,
		oauth.ErrorKindInvalidConfig,
		"create client",
		cause,
	)
}
