package cryptoutil

import (
	"encoding/base64"
)

func Base64Encode(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}

func Base64EncodeString(s string) string {
	return Base64Encode([]byte(s))
}

func Base64Decode(s string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(s)
}
