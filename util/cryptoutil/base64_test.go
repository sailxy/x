package cryptoutil

import (
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBase64(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		want string
	}{
		{
			name: "hello",
			data: []byte("hello"),
			want: base64.StdEncoding.EncodeToString([]byte("hello")),
		},
		{
			name: "empty",
			data: []byte(""),
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, Base64Encode(tt.data))
			assert.Equal(t, tt.want, Base64EncodeString(string(tt.data)))

			got, err := Base64Decode(tt.want)
			assert.NoError(t, err)
			assert.Equal(t, tt.data, got)
		})
	}
}

func TestBase64Decode(t *testing.T) {
	_, err := Base64Decode("%%%")
	assert.Error(t, err)
}
