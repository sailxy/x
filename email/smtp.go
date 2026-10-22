package email

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"time"

	mail "github.com/wneessen/go-mail"
)

// TLSMode controls SMTP transport independently of the configured port.
type TLSMode string

const (
	// TLSModeAuto uses implicit TLS on port 465 and opportunistic STARTTLS otherwise.
	TLSModeAuto     TLSMode = "auto"
	TLSModeImplicit TLSMode = "implicit"
	TLSModeSTARTTLS TLSMode = "starttls"
)

type Config struct {
	Host     string
	Port     int
	Username string
	Password string
	TLSMode  TLSMode       // Empty defaults to TLSModeAuto.
	Timeout  time.Duration // Zero defaults to 30 seconds; negative values are invalid.
}

type SMTPDialer struct {
	config    Config
	tlsConfig *tls.Config
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
	if c.TLSMode == "" {
		c.TLSMode = TLSModeAuto
	}
	if c.Timeout == 0 {
		c.Timeout = 30 * time.Second
	}
	return &SMTPDialer{config: c, tlsConfig: &tls.Config{ServerName: c.Host, MinVersion: tls.VersionTLS12}}
}

// Send bounds SMTP delivery to the configured timeout and closes the connection on cancellation.
func (s *SMTPDialer) Send(ctx context.Context, c SendConfig) error {
	if s.config.Timeout <= 0 {
		return errors.New("smtp timeout must be positive")
	}
	implicit := false
	policy := mail.TLSOpportunistic
	switch s.config.TLSMode {
	case TLSModeAuto:
		implicit = s.config.Port == 465
	case TLSModeImplicit:
		implicit = true
	case TLSModeSTARTTLS:
		policy = mail.TLSMandatory
	default:
		return errors.New("invalid smtp TLS mode")
	}
	ctx, cancel := context.WithTimeout(ctx, s.config.Timeout)
	defer cancel()
	if err := ctx.Err(); err != nil {
		return err
	}

	message := mail.NewMsg()
	if err := message.FromFormat(c.FromName, c.FromAddr); err != nil {
		return err
	}
	if err := message.AddToFormat(c.ToName, c.ToAddr); err != nil {
		return err
	}
	if c.CcAddr != "" {
		if err := message.AddCcFormat(c.CcName, c.CcAddr); err != nil {
			return err
		}
	}
	message.Subject(c.Subject)
	message.SetBodyString(mail.TypeTextHTML, c.Body)
	if c.Attach != "" {
		file, err := os.Open(c.Attach)
		if err != nil {
			return fmt.Errorf("open email attachment: %w", err)
		}
		defer file.Close()
		message.AttachReadSeeker(filepath.Base(c.Attach), file)
	}

	// go-mail's context bounds dialing; this hook also cancels in-flight SMTP reads/writes.
	var cleanup func()
	defer func() {
		if cleanup != nil {
			cleanup()
		}
	}()
	opts := []mail.Option{
		mail.WithPort(s.config.Port), mail.WithTimeout(s.config.Timeout),
		mail.WithTLSPolicy(policy), mail.WithTLSConfig(s.tlsConfig),
		mail.WithDialContextFunc(func(dialCtx context.Context, network, address string) (net.Conn, error) {
			conn, err := (&net.Dialer{}).DialContext(dialCtx, network, address)
			if err != nil {
				return nil, err
			}
			stop := context.AfterFunc(ctx, func() { conn.Close() })
			cleanup = func() { stop(); conn.Close() }
			if implicit {
				return tls.Client(conn, s.tlsConfig), nil
			}
			return conn, nil
		}),
	}
	if implicit {
		opts = append(opts, mail.WithSSL())
	}
	if s.config.Username != "" {
		opts = append(opts, mail.WithUsername(s.config.Username), mail.WithPassword(s.config.Password),
			mail.WithOpportunisticSMTPAuth(mail.SMTPAuthCramMD5, mail.SMTPAuthPlain, mail.SMTPAuthLogin))
	}
	client, err := mail.NewClient(s.config.Host, opts...)
	if err != nil {
		return err
	}
	err = client.DialAndSendWithContext(ctx, message)
	if err != nil {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if deadline, ok := ctx.Deadline(); ok && !time.Now().Before(deadline) {
			return context.DeadlineExceeded
		}
	}
	return err
}
