package password

import (
	"golang.org/x/crypto/bcrypt"
)

// Encrypt passwords using the bcrypt algorithm.
func Encrypt(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// Check if the password is correct.
func Check(hash, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}
