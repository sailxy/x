package oauth

import (
	"encoding/json"
	"fmt"
	"io"
)

const redactedValue = "[REDACTED]"

// SensitiveString holds a credential that callers need to use but that must
// not be exposed through ordinary formatting or JSON encoding.
type SensitiveString string

// Value returns the original credential for immediate protocol use.
func (s SensitiveString) Value() string {
	return string(s)
}

func (SensitiveString) String() string {
	return redactedValue
}

func (SensitiveString) GoString() string {
	return redactedValue
}

// Format redacts every fmt verb, including verbs that would otherwise format
// the underlying string directly.
func (SensitiveString) Format(state fmt.State, _ rune) {
	_, _ = io.WriteString(state, redactedValue)
}

func (SensitiveString) MarshalJSON() ([]byte, error) {
	return json.Marshal(redactedValue)
}

func (SensitiveString) MarshalText() ([]byte, error) {
	return []byte(redactedValue), nil
}
