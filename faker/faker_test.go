package faker

import "testing"

func TestEmail(t *testing.T) {
	email := Email()
	t.Log(email)
}

func TestPassword(t *testing.T) {
	password := Password()
	t.Log(password)
}

func TestPhone(t *testing.T) {
	phone := Phone()
	t.Log(phone)
}
