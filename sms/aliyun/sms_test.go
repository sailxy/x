package sms

import "testing"

func TestNewReturnsError(t *testing.T) {
	client, err := New(Config{})
	if err == nil || client != nil {
		t.Fatal("New should return an error for incomplete configuration")
	}
}
