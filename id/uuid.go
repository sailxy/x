package id

import (
	"github.com/gofrs/uuid/v5"
)

// NewUUID returns a new UUID v7.
func NewUUID() (string, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return "", err
	}
	return id.String(), nil
}
