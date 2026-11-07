package mail

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log/slog"
	"mime"
	"net"
	"net/mail"
	"net/smtp"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Sender interface {
	SendVerificationCode(ctx context.Context, recipient, code string) error
}

type LogSender struct{ Log *slog.Logger }

func (s LogSender) SendVerificationCode(ctx context.Context, recipient, code string) error {
	s.Log.InfoContext(ctx, "development email verification code", "recipient", recipient, "code", code)
	return nil
}

type SMTPConfig struct {
	Address, Username, Password, From string
	TLSMode                           string
}

type SMTPSender struct {
	Config  SMTPConfig
	rootCAs *x509.CertPool
}

func (s SMTPSender) SendVerificationCode(ctx context.Context, recipient, code string) error {
	return s.withClient(ctx, func(client *smtp.Client) error { return s.sendMessage(client, recipient, code) })
}

// Check verifies TLS and SMTP credentials without delivering an email.
func (s SMTPSender) Check(ctx context.Context) error {
	return s.withClient(ctx, func(_ *smtp.Client) error { return nil })
}

func (s SMTPSender) withClient(ctx context.Context, use func(*smtp.Client) error) error {
	host, _, err := net.SplitHostPort(s.Config.Address)
	if err != nil {
		return fmt.Errorf("parse SMTP address: %w", err)
	}
	conn, err := (&net.Dialer{Timeout: 10 * time.Second}).DialContext(ctx, "tcp", s.Config.Address)
	if err != nil {
		return fmt.Errorf("connect SMTP: %w", err)
	}
	defer conn.Close()
	deadline := time.Now().Add(20 * time.Second)
	if ctxDeadline, ok := ctx.Deadline(); ok && ctxDeadline.Before(deadline) {
		deadline = ctxDeadline
	}
	if err := conn.SetDeadline(deadline); err != nil {
		return fmt.Errorf("set SMTP deadline: %w", err)
	}
	rawConn := conn
	stopCancel := context.AfterFunc(ctx, func() { _ = rawConn.Close() })
	defer stopCancel()
	tlsConfig := &tls.Config{ServerName: host, MinVersion: tls.VersionTLS12, RootCAs: s.rootCAs}
	_, port, _ := net.SplitHostPort(s.Config.Address)
	mode := s.Config.TLSMode
	if mode == "" || mode == "auto" {
		if port == "465" {
			mode = "tls"
		} else {
			mode = "starttls"
		}
	}
	if mode != "tls" && mode != "starttls" {
		return fmt.Errorf("invalid SMTP TLS mode")
	}
	if mode == "tls" {
		tlsConn := tls.Client(conn, tlsConfig)
		if err := tlsConn.HandshakeContext(ctx); err != nil {
			return fmt.Errorf("connect SMTP TLS: %w", err)
		}
		conn = tlsConn
	}
	client, err := smtp.NewClient(conn, host)
	if err != nil {
		return fmt.Errorf("create SMTP client: %w", err)
	}
	defer client.Close()
	if mode == "starttls" {
		if ok, _ := client.Extension("STARTTLS"); !ok {
			return fmt.Errorf("SMTP server does not support STARTTLS")
		} else if err := client.StartTLS(tlsConfig); err != nil {
			return fmt.Errorf("start SMTP TLS: %w", err)
		}
	}
	if s.Config.Username != "" {
		if err := client.Auth(smtp.PlainAuth("", s.Config.Username, s.Config.Password, host)); err != nil {
			return fmt.Errorf("authenticate SMTP: %w", err)
		}
	}
	if err := use(client); err != nil {
		return err
	}
	if err := client.Quit(); err != nil {
		return fmt.Errorf("quit SMTP: %w", err)
	}
	return nil
}

func (s SMTPSender) sendMessage(client *smtp.Client, recipient, code string) error {
	from, err := mail.ParseAddress(s.Config.From)
	if err != nil {
		return fmt.Errorf("invalid sender mailbox")
	}
	_, domain, ok := strings.Cut(from.Address, "@")
	if !ok {
		return fmt.Errorf("invalid sender domain")
	}
	if err := client.Mail(s.Config.From); err != nil {
		return fmt.Errorf("SMTP MAIL FROM: %w", err)
	}
	if err := client.Rcpt(recipient); err != nil {
		return fmt.Errorf("SMTP RCPT TO: %w", err)
	}
	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("SMTP DATA: %w", err)
	}
	body := strings.Join([]string{
		"From: " + s.Config.From,
		"To: " + recipient,
		"Subject: " + mime.QEncoding.Encode("utf-8", "Код подтверждения почты — КлёвЛига"),
		"Date: " + time.Now().Format(time.RFC1123Z),
		"Message-ID: <" + uuid.NewString() + "@" + domain + ">",
		"MIME-Version: 1.0",
		"Content-Type: text/plain; charset=UTF-8",
		"",
		"Код подтверждения КлёвЛиги: " + code,
		"Код действует 10 минут. Если вы не создавали аккаунт, проигнорируйте письмо.",
	}, "\r\n")
	if _, err := w.Write([]byte(body)); err != nil {
		return fmt.Errorf("write SMTP message: %w", err)
	}
	if err := w.Close(); err != nil {
		return fmt.Errorf("send SMTP message: %w", err)
	}
	return nil
}
