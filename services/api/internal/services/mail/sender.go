package mail

import (
	"context"
	"crypto/tls"
	"fmt"
	"log/slog"
	"net"
	"net/smtp"
	"strings"
	"time"
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
}

type SMTPSender struct{ Config SMTPConfig }

func (s SMTPSender) SendVerificationCode(ctx context.Context, recipient, code string) error {
	host, _, err := net.SplitHostPort(s.Config.Address)
	if err != nil {
		return fmt.Errorf("parse SMTP address: %w", err)
	}
	conn, err := (&net.Dialer{Timeout: 10 * time.Second}).DialContext(ctx, "tcp", s.Config.Address)
	if err != nil {
		return fmt.Errorf("connect SMTP: %w", err)
	}
	client, err := smtp.NewClient(conn, host)
	if err != nil {
		return fmt.Errorf("create SMTP client: %w", err)
	}
	defer client.Close()
	if ok, _ := client.Extension("STARTTLS"); !ok {
		return fmt.Errorf("SMTP server does not support STARTTLS")
	} else if err := client.StartTLS(&tls.Config{ServerName: host, MinVersion: tls.VersionTLS12}); err != nil {
		return fmt.Errorf("start SMTP TLS: %w", err)
	}
	if s.Config.Username != "" {
		if err := client.Auth(smtp.PlainAuth("", s.Config.Username, s.Config.Password, host)); err != nil {
			return fmt.Errorf("authenticate SMTP: %w", err)
		}
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
		"Subject: =?UTF-8?B?0JrQvtC0INC/0L7QtNGC0LLQtdGA0LbQtNC10L3QuNGPINCf0L7Rh9GC0Ys=?=",
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
	if err := client.Quit(); err != nil {
		return fmt.Errorf("quit SMTP: %w", err)
	}
	return nil
}
