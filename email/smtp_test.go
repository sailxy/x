package email

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSend(t *testing.T) {
	smtpDialer := NewSMTPDialer(Config{
		Host:     "smtp.qiye.aliyun.com",
		Port:     465,
		Username: "name@example.com",
		Password: "password",
	})
	err := smtpDialer.Send(SendConfig{
		FromAddr: "name@example.com",
		FromName: "From",
		ToAddr:   "to@example.com",
		ToName:   "To",
		CcAddr:   "cc@example.com",
		CcName:   "Cc",
		Subject:  "Hello",
		Body:     "This is a test email.",
		Attach:   "/tmp/1.txt",
	})
	assert.NoError(t, err)
}
