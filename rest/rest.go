package rest

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"resty.dev/v3"
)

const (
	// DefaultTimeout bounds requests when the caller does not provide an HTTP
	// client.
	DefaultTimeout = 10 * time.Second

	// MaxResponseBodyBytes bounds response bodies held in memory.
	MaxResponseBodyBytes int64 = 1 << 20
)

// ErrResponseTooLarge indicates that a response exceeded
// MaxResponseBodyBytes.
var ErrResponseTooLarge = errors.New("REST response body is too large")

// Response contains the bounded parts of an HTTP response needed by callers.
// Body is available for both successful and unsuccessful statuses.
type Response struct {
	StatusCode int
	Header     http.Header
	Body       []byte
}

// Format omits Header and Body so direct logging cannot expose platform
// credentials contained in a response.
func (r Response) Format(state fmt.State, _ rune) {
	_, _ = fmt.Fprintf(state, "REST response status %d (%d body bytes)", r.StatusCode, len(r.Body))
}

// StatusError reports an unsuccessful HTTP status without retaining or
// formatting the response body.
type StatusError struct {
	StatusCode int
}

func (e *StatusError) Error() string {
	return fmt.Sprintf("REST request returned status %d", e.StatusCode)
}

type REST struct {
	client    *resty.Client
	doTimeout time.Duration
}

func NewREST() *REST {
	return &REST{
		client:    resty.New(),
		doTimeout: DefaultTimeout,
	}
}

// NewRESTWithClient creates a REST client backed by client. When client is nil,
// it is equivalent to NewREST. Do always applies DefaultTimeout; a shorter
// timeout configured on client still takes precedence.
func NewRESTWithClient(client *http.Client) *REST {
	if client == nil {
		return NewREST()
	}
	return &REST{
		client:    resty.NewWithClient(client),
		doTimeout: DefaultTimeout,
	}
}

func (r *REST) Close() error {
	return r.client.Close()
}

// Get sends a GET request.
func (r *REST) Get(ctx context.Context, address string) (*Response, error) {
	return r.request(ctx, http.MethodGet, address, nil)
}

// Post sends a POST request with body.
func (r *REST) Post(ctx context.Context, address string, body io.Reader) (*Response, error) {
	return r.request(ctx, http.MethodPost, address, body)
}

func (r *REST) request(ctx context.Context, method, address string, body io.Reader) (*Response, error) {
	req, err := http.NewRequest(method, address, body)
	if err != nil {
		return nil, fmt.Errorf("create REST request: %w", err)
	}
	return r.Do(ctx, req)
}

// Do executes req with ctx, reads at most MaxResponseBodyBytes, and rejects
// non-2xx responses. A bounded Response is returned with StatusError so callers
// can safely decode a provider-specific error code.
func (r *REST) Do(ctx context.Context, req *http.Request) (*Response, error) {
	if ctx == nil {
		return nil, errors.New("REST context is nil")
	}
	if req == nil {
		return nil, errors.New("REST request is nil")
	}
	if r.doTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, r.doTimeout)
		defer cancel()
	}

	resp, err := r.client.Client().Do(req.Clone(ctx))
	if err != nil {
		return nil, fmt.Errorf("execute REST request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, MaxResponseBodyBytes+1))
	if err != nil {
		return nil, fmt.Errorf("read REST response: %w", err)
	}
	if int64(len(body)) > MaxResponseBodyBytes {
		return nil, ErrResponseTooLarge
	}

	result := &Response{
		StatusCode: resp.StatusCode,
		Header:     resp.Header.Clone(),
		Body:       body,
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return result, &StatusError{StatusCode: resp.StatusCode}
	}
	return result, nil
}
