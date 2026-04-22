package cryptoutil

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
)

func MD5(data []byte) string {
	sum := md5.Sum(data)
	return hex.EncodeToString(sum[:])
}

func MD5String(s string) string {
	return MD5([]byte(s))
}

func SHA256(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func SHA256String(s string) string {
	return SHA256([]byte(s))
}
