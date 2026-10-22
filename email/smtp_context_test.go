package email

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestSend_ClosesStalledConnection(t *testing.T) {
	for _, stage := range []string{"greeting", "RCPT", "DATA", "TLS", "cancel"} {
		t.Run(stage, func(t *testing.T) {
			listener, err := net.Listen("tcp", "127.0.0.1:0")
			require.NoError(t, err)
			defer listener.Close()
			closed := make(chan error, 1)
			stalled := make(chan struct{})
			go func() {
				conn, err := listener.Accept()
				if err != nil {
					closed <- err
					return
				}
				defer conn.Close()
				if err := conn.SetDeadline(time.Now().Add(3 * time.Second)); err != nil {
					closed <- err
					return
				}
				protocol := textproto.NewConn(conn)
				stall := func() {
					close(stalled)
					_, err := io.Copy(io.Discard, conn)
					closed <- err
				}
				if stage == "greeting" || stage == "TLS" || stage == "cancel" {
					stall()
					return
				}
				if err := protocol.PrintfLine("220 localhost test SMTP"); err != nil {
					closed <- err
					return
				}
				for {
					line, err := protocol.ReadLine()
					if err != nil {
						closed <- err
						return
					}
					if strings.HasPrefix(line, stage) {
						stall()
						return
					}
					if err := protocol.PrintfLine("250 localhost"); err != nil {
						closed <- err
						return
					}
				}
			}()
			host, port, err := net.SplitHostPort(listener.Addr().String())
			require.NoError(t, err)
			number, err := strconv.Atoi(port)
			require.NoError(t, err)
			mode := TLSModeAuto
			if stage == "TLS" {
				mode = TLSModeImplicit
			}
			sender := NewSMTPDialer(Config{Host: host, Port: number, TLSMode: mode, Timeout: 200 * time.Millisecond})
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			if stage == "cancel" {
				go func() { <-stalled; cancel() }()
			}
			started := time.Now()
			err = sender.Send(ctx, SendConfig{FromAddr: "sender@example.test", ToAddr: "to@example.test", Subject: "test", Body: "test"})
			if stage == "cancel" {
				require.ErrorIs(t, err, context.Canceled)
			} else {
				require.ErrorIs(t, err, context.DeadlineExceeded)
			}
			require.Less(t, time.Since(started), 2*time.Second)
			select {
			case err := <-closed:
				require.NoError(t, err)
			case <-time.After(time.Second):
				t.Fatal("SMTP connection remained open after cancellation")
			}
		})
	}
}

func TestSend_LOGINAndAttachment(t *testing.T) {
	fixture := httptest.NewTLSServer(http.NotFoundHandler())
	certificate := fixture.TLS.Certificates[0]
	fixture.Close()
	roots := x509.NewCertPool()
	leaf, err := x509.ParseCertificate(certificate.Certificate[0])
	require.NoError(t, err)
	roots.AddCert(leaf)
	attachment := filepath.Join(t.TempDir(), "proof.txt")
	require.NoError(t, os.WriteFile(attachment, []byte("attachment content"), 0600))
	for _, implicit := range []bool{true, false} {
		t.Run(fmt.Sprintf("implicitTLS=%t", implicit), func(t *testing.T) {
			listener, err := net.Listen("tcp", "127.0.0.1:0")
			require.NoError(t, err)
			defer listener.Close()
			result := make(chan error, 1)
			message := make(chan []byte, 1)
			go func() {
				result <- loginSMTP(listener, certificate, implicit, message)
			}()
			host, port, err := net.SplitHostPort(listener.Addr().String())
			require.NoError(t, err)
			number, err := strconv.Atoi(port)
			require.NoError(t, err)
			mode := TLSModeSTARTTLS
			if implicit {
				mode = TLSModeImplicit
			}
			sender := NewSMTPDialer(Config{Host: host, Port: number, Username: "sender@example.test", Password: "test-password", TLSMode: mode})
			sender.tlsConfig.RootCAs = roots
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			require.NoError(t, sender.Send(ctx, SendConfig{
				FromAddr: "sender@example.test", FromName: "Sender", ToAddr: "to@example.test", ToName: "Recipient",
				CcAddr: "copy@example.test", CcName: "Copy", Subject: "Test subject", Body: "<p>Test body</p>", Attach: attachment,
			}))
			require.NoError(t, <-result)
			raw := string(<-message)
			for _, text := range []string{"sender@example.test", "to@example.test", "copy@example.test", "Subject: Test subject", "<p>Test body</p>", `attachment; filename="proof.txt"`, "YXR0YWNobWVudCBjb250ZW50"} {
				require.Contains(t, raw, text)
			}
		})
	}
}

func loginSMTP(listener net.Listener, certificate tls.Certificate, implicit bool, message chan<- []byte) error {
	conn, err := listener.Accept()
	if err != nil {
		return err
	}
	defer conn.Close()
	if err := conn.SetDeadline(time.Now().Add(3 * time.Second)); err != nil {
		return err
	}
	tlsConfig := &tls.Config{Certificates: []tls.Certificate{certificate}, MinVersion: tls.VersionTLS12}
	encrypted := implicit
	if implicit {
		conn = tls.Server(conn, tlsConfig)
	}
	protocol := textproto.NewConn(conn)
	if err := protocol.PrintfLine("220 localhost test SMTP"); err != nil {
		return err
	}
	for {
		line, err := protocol.ReadLine()
		if err != nil {
			return err
		}
		reply := "250 OK"
		switch {
		case strings.HasPrefix(line, "EHLO"):
			reply = "250-localhost\r\n250 AUTH LOGIN"
			if !encrypted {
				reply = "250-localhost\r\n250-STARTTLS\r\n250 AUTH LOGIN"
			}
		case line == "STARTTLS":
			if err := protocol.PrintfLine("220 start TLS"); err != nil {
				return err
			}
			conn = tls.Server(conn, tlsConfig)
			protocol = textproto.NewConn(conn)
			encrypted = true
			continue
		case line == "AUTH LOGIN":
			for _, step := range []struct{ challenge, expected string }{{"Username:", "sender@example.test"}, {"Password:", "test-password"}} {
				if err := protocol.PrintfLine("334 %s", base64.StdEncoding.EncodeToString([]byte(step.challenge))); err != nil {
					return err
				}
				answer, err := protocol.ReadLine()
				if err != nil {
					return err
				}
				decoded, err := base64.StdEncoding.DecodeString(answer)
				if err != nil || string(decoded) != step.expected {
					return fmt.Errorf("unexpected SMTP authentication response")
				}
			}
			reply = "235 authenticated"
		case line == "DATA":
			if err := protocol.PrintfLine("354 send data"); err != nil {
				return err
			}
			raw, err := protocol.ReadDotBytes()
			if err != nil {
				return err
			}
			message <- raw
		case line == "QUIT":
			return protocol.PrintfLine("221 bye")
		}
		if err := protocol.PrintfLine("%s", reply); err != nil {
			return err
		}
	}
}

func TestSend_InvalidConfig(t *testing.T) {
	for _, tt := range []struct {
		config  Config
		message string
	}{
		{Config{TLSMode: "unknown"}, "invalid smtp TLS mode"},
		{Config{Timeout: -time.Second}, "smtp timeout must be positive"},
	} {
		require.ErrorContains(t, NewSMTPDialer(tt.config).Send(context.Background(), SendConfig{}), tt.message)
	}
	defaults := NewSMTPDialer(Config{})
	require.Equal(t, TLSModeAuto, defaults.config.TLSMode)
	require.Equal(t, 30*time.Second, defaults.config.Timeout)
}

func TestSend_MissingAttachmentDoesNotConnect(t *testing.T) {
	listener, err := net.ListenTCP("tcp", &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1)})
	require.NoError(t, err)
	defer listener.Close()
	sender := NewSMTPDialer(Config{Host: "127.0.0.1", Port: listener.Addr().(*net.TCPAddr).Port, Timeout: 200 * time.Millisecond})
	err = sender.Send(context.Background(), SendConfig{
		FromAddr: "sender@example.test", ToAddr: "to@example.test", Subject: "test", Body: "test",
		Attach: filepath.Join(t.TempDir(), "missing.txt"),
	})
	require.ErrorIs(t, err, os.ErrNotExist)
	require.NoError(t, listener.SetDeadline(time.Now().Add(50*time.Millisecond)))
	conn, err := listener.Accept()
	if conn != nil {
		conn.Close()
		t.Fatal("invalid attachment reached SMTP")
	}
	var networkError net.Error
	require.ErrorAs(t, err, &networkError)
	require.True(t, networkError.Timeout())
}
