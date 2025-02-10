package id

import gonanoid "github.com/matoous/go-nanoid/v2"

func NewNanoID(length int) (string, error) {
	return gonanoid.Generate("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789", length)
}
