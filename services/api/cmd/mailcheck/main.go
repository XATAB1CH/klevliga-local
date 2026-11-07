// mailcheck checks SMTP TLS and authentication without sending messages.
package main

import (
	"context"
	"flag"
	"fmt"
	"net/mail"
	"os"
	"time"

	"github.com/XATAB1CH/klev-liga/internal/config"
	mailSvc "github.com/XATAB1CH/klev-liga/internal/services/mail"
)

func main() {
	path := flag.String("c", "./config.yaml", "API configuration file")
	sendTestTo := flag.String("send-test-to", "", "explicitly send one test email to this address")
	flag.Parse()
	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
	defer cancel()
	cfg, err := config.New(ctx, *path)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Cannot load configuration. Check config.yaml placeholders and .env syntax.")
		os.Exit(1)
	}
	if cfg.Mail.Mode != "smtp" || cfg.Mail.Username == "" || cfg.Mail.Password == "" || cfg.Mail.Password == "..." {
		fmt.Fprintln(os.Stderr, "Set MAIL_MODE=smtp and fill SMTP_USERNAME / SMTP_PASSWORD in .env. No email sent.")
		os.Exit(1)
	}
	sender := mailSvc.SMTPSender{Config: mailSvc.SMTPConfig{Address: cfg.Mail.Address, TLSMode: cfg.Mail.TLSMode, Username: cfg.Mail.Username, Password: cfg.Mail.Password, From: cfg.Mail.From}}
	if *sendTestTo != "" {
		address, err := mail.ParseAddress(*sendTestTo)
		if err != nil || address.Name != "" || address.Address != *sendTestTo {
			fmt.Fprintln(os.Stderr, "Invalid test recipient address.")
			os.Exit(1)
		}
		if err := sender.SendVerificationCode(ctx, address.Address, "123456"); err != nil {
			fmt.Fprintln(os.Stderr, "Test email failed:", err)
			os.Exit(1)
		}
		fmt.Println("One test email accepted by SMTP. Code 123456 is diagnostic only; no account was created.")
		return
	}
	if err := sender.Check(ctx); err != nil {
		fmt.Fprintln(os.Stderr, "SMTP check failed:", err)
		os.Exit(1)
	}
	fmt.Println("SMTP TLS connection and authentication succeeded. No email sent.")
}
