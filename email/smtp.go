package email

import "gopkg.in/gomail.v2"

type Config struct {
	Host     string
	Port     int
	Username string
	Password string
}

type SMTPDialer struct {
	dialer *gomail.Dialer
}

type SendConfig struct {
	FromAddr string
	FromName string
	ToAddr   string
	ToName   string
	CcAddr   string
	CcName   string
	Subject  string
	Body     string
	Attach   string
}

func NewSMTPDialer(c Config) *SMTPDialer {
	return &SMTPDialer{
		dialer: gomail.NewDialer(c.Host, c.Port, c.Username, c.Password),
	}
}

func (s *SMTPDialer) Send(c SendConfig) error {
	m := gomail.NewMessage()
	m.SetAddressHeader("From", c.FromAddr, c.FromName)
	m.SetAddressHeader("To", c.ToAddr, c.ToName)
	if c.CcAddr != "" {
		m.SetAddressHeader("Cc", c.CcAddr, c.CcName)
	}
	m.SetHeader("Subject", c.Subject)
	m.SetBody("text/html", c.Body)
	if c.Attach != "" {
		m.Attach(c.Attach)
	}

	return s.dialer.DialAndSend(m)
}
