// Package oauth contains shared primitives for OAuth and OpenID Connect
// provider clients.
package oauth

import (
	"errors"
	"fmt"
)

// Provider identifies the remote authentication provider.
type Provider string

const (
	ProviderGitHub Provider = "github"
	ProviderApple  Provider = "apple"
)

// ErrorKind identifies a stable OAuth client failure category.
type ErrorKind string

const (
	ErrorKindInvalidConfig   ErrorKind = "invalid_config"
	ErrorKindInvalidInput    ErrorKind = "invalid_input"
	ErrorKindTransport       ErrorKind = "transport"
	ErrorKindPlatform        ErrorKind = "platform_rejected"
	ErrorKindDecode          ErrorKind = "decode"
	ErrorKindInvalidResponse ErrorKind = "invalid_response"
	ErrorKindTokenValidation ErrorKind = "token_validation"
)

var (
	ErrInvalidConfig   = &Error{Kind: ErrorKindInvalidConfig}
	ErrInvalidInput    = &Error{Kind: ErrorKindInvalidInput}
	ErrTransport       = &Error{Kind: ErrorKindTransport}
	ErrPlatform        = &Error{Kind: ErrorKindPlatform}
	ErrDecode          = &Error{Kind: ErrorKindDecode}
	ErrInvalidResponse = &Error{Kind: ErrorKindInvalidResponse}
	ErrTokenValidation = &Error{Kind: ErrorKindTokenValidation}
)

// Error contains only metadata safe for callers to inspect. The wrapped cause
// and provider error code are intentionally omitted from formatted text.
type Error struct {
	Provider   Provider
	Kind       ErrorKind
	Operation  string
	StatusCode int
	Code       string

	cause error
}

// NewError creates a classified provider error and retains cause for
// errors.Is and errors.As without formatting the cause text.
func NewError(provider Provider, kind ErrorKind, operation string, cause error) *Error {
	return &Error{
		Provider:  provider,
		Kind:      kind,
		Operation: operation,
		cause:     cause,
	}
}

func (e *Error) Error() string {
	message := "oauth: " + string(e.Kind)
	if e.Provider != "" {
		message = "oauth " + string(e.Provider) + ": " + string(e.Kind)
	}
	if e.Operation != "" {
		message = "oauth " + string(e.Provider) + " " + e.Operation + ": " + string(e.Kind)
	}
	if e.StatusCode != 0 {
		message += fmt.Sprintf(" (status %d)", e.StatusCode)
	}
	return message
}

// GoString keeps the wrapped cause and provider error text out of Go-syntax
// formatting such as %#v.
func (e *Error) GoString() string {
	return e.Error()
}

// Unwrap exposes the underlying cause without including it in Error().
func (e *Error) Unwrap() error {
	return e.cause
}

// Is matches OAuth errors by stable failure category.
func (e *Error) Is(target error) bool {
	var targetError *Error
	return errors.As(target, &targetError) && e.Kind == targetError.Kind
}
