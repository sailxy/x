package oauth

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSensitiveStringRedaction(t *testing.T) {
	const raw = "oauth-access-token-fixture"
	secret := SensitiveString(raw)
	assert.Equal(t, raw, secret.Value())

	for _, format := range []string{"%s", "%q", "%v", "%+v", "%#v", "%x"} {
		formatted := fmt.Sprintf(format, secret)
		assert.NotContains(t, formatted, raw)
		assert.Contains(t, formatted, redactedValue)
	}

	encoded, err := json.Marshal(struct {
		Token SensitiveString `json:"token"`
	}{Token: secret})
	require.NoError(t, err)
	assert.NotContains(t, string(encoded), raw)
	assert.Contains(t, string(encoded), redactedValue)
}
