package oauth

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestErrorClassification(t *testing.T) {
	tests := []struct {
		name   string
		kind   ErrorKind
		target error
	}{
		{"invalid config", ErrorKindInvalidConfig, ErrInvalidConfig},
		{"invalid input", ErrorKindInvalidInput, ErrInvalidInput},
		{"transport", ErrorKindTransport, ErrTransport},
		{"platform", ErrorKindPlatform, ErrPlatform},
		{"decode", ErrorKindDecode, ErrDecode},
		{"invalid response", ErrorKindInvalidResponse, ErrInvalidResponse},
		{"token validation", ErrorKindTokenValidation, ErrTokenValidation},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewError(ProviderGitHub, tt.kind, "test", nil)
			assert.ErrorIs(t, err, tt.target)
			unrelated := error(ErrInvalidInput)
			if tt.kind == ErrorKindInvalidInput {
				unrelated = ErrInvalidConfig
			}
			assert.NotErrorIs(t, err, unrelated)

			var oauthError *Error
			require.ErrorAs(t, err, &oauthError)
			assert.Equal(t, ProviderGitHub, oauthError.Provider)
			assert.Equal(t, tt.kind, oauthError.Kind)
			assert.Equal(t, "test", oauthError.Operation)
		})
	}
}

func TestErrorUnwrapsCause(t *testing.T) {
	err := NewError(ProviderApple, ErrorKindTransport, "exchange code", context.Canceled)
	assert.ErrorIs(t, err, ErrTransport)
	assert.ErrorIs(t, err, context.Canceled)

	var oauthError *Error
	require.True(t, errors.As(err, &oauthError))
	assert.Equal(t, ProviderApple, oauthError.Provider)
}

func TestErrorFormattingRedactsSensitiveValues(t *testing.T) {
	const raw = "provider-token-fixture"
	for _, provider := range []Provider{ProviderGitHub, ProviderApple} {
		for _, kind := range []ErrorKind{
			ErrorKindInvalidConfig,
			ErrorKindInvalidInput,
			ErrorKindTransport,
			ErrorKindPlatform,
			ErrorKindDecode,
			ErrorKindInvalidResponse,
			ErrorKindTokenValidation,
		} {
			err := NewError(provider, kind, "verify token", fmt.Errorf("remote response contained %s", raw))
			err.Code = raw
			err.StatusCode = 401
			for _, format := range []string{"%s", "%q", "%v", "%+v", "%#v", "%x"} {
				assert.NotContains(t, fmt.Sprintf(format, err), raw)
			}
		}
	}
}
