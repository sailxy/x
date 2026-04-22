package cryptoutil

import (
	"crypto/hmac"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"hash"
)

func HMACSHA256(data, secret []byte) string {
	return hex.EncodeToString(hmacSum(sha256.New, data, secret))
}

func HMACSHA256String(data, secret string) string {
	return HMACSHA256([]byte(data), []byte(secret))
}

func HMACSHA1Base64(data, secret []byte) string {
	return base64.StdEncoding.EncodeToString(hmacSum(sha1.New, data, secret))
}

func HMACSHA1Base64String(data, secret string) string {
	return HMACSHA1Base64([]byte(data), []byte(secret))
}

func VerifyHMACSHA256(data, secret []byte, signature string) (bool, error) {
	want, err := hex.DecodeString(signature)
	if err != nil {
		return false, fmt.Errorf("decode hmac signature: %w", err)
	}

	return hmac.Equal(hmacSum(sha256.New, data, secret), want), nil
}

func VerifyHMACSHA256String(data, secret, signature string) (bool, error) {
	return VerifyHMACSHA256([]byte(data), []byte(secret), signature)
}

func hmacSum(newHash func() hash.Hash, data, secret []byte) []byte {
	mac := hmac.New(newHash, secret)
	_, _ = mac.Write(data)
	return mac.Sum(nil)
}
