package mail

import (
	"bufio"
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"net"
	"net/http/httptest"
	"net/mail"
	"net/textproto"
	"strings"
	"testing"
	"time"
)

func TestSMTPVerificationOverTLSAndSTARTTLS(t *testing.T) {
	certificateServer := httptest.NewTLSServer(nil)
	certificates := certificateServer.TLS.Certificates
	roots := x509.NewCertPool()
	roots.AddCert(certificateServer.Certificate())
	certificateServer.Close()
	for _, testCase := range []struct {
		mode      string
		checkOnly bool
	}{{"tls", false}, {"starttls", false}, {"auto", false}, {"tls", true}} {
		mode := testCase.mode
		t.Run(fmt.Sprintf("%s/check=%t", mode, testCase.checkOnly), func(t *testing.T) {
			listener, err := net.Listen("tcp", "127.0.0.1:0")
			if err != nil {
				t.Fatal(err)
			}
			defer listener.Close()
			serverTLS := &tls.Config{Certificates: certificates, MinVersion: tls.VersionTLS12}
			if mode == "tls" {
				listener = tls.NewListener(listener, serverTLS)
			}
			done := make(chan error, 1)
			go func() {
				conn, err := listener.Accept()
				if err != nil {
					done <- err
					return
				}
				defer conn.Close()
				_ = conn.SetDeadline(time.Now().Add(5 * time.Second))
				secured := mode == "tls"
				reader := bufio.NewReader(conn)
				_, _ = fmt.Fprint(conn, "220 localhost test SMTP\r\n")
				var message string
				for {
					line, err := reader.ReadString('\n')
					if err != nil {
						done <- err
						return
					}
					switch {
					case strings.HasPrefix(line, "EHLO"):
						if secured {
							_, _ = fmt.Fprint(conn, "250-localhost\r\n250 AUTH PLAIN\r\n")
						} else {
							_, _ = fmt.Fprint(conn, "250-localhost\r\n250 STARTTLS\r\n")
						}
					case strings.HasPrefix(line, "STARTTLS"):
						_, _ = fmt.Fprint(conn, "220 begin TLS\r\n")
						tlsConn := tls.Server(conn, serverTLS)
						if err := tlsConn.Handshake(); err != nil {
							done <- err
							return
						}
						conn = tlsConn
						reader = bufio.NewReader(conn)
						secured = true
					case strings.HasPrefix(line, "AUTH"):
						if !secured {
							done <- fmt.Errorf("credentials were sent without TLS")
							return
						}
						_, _ = fmt.Fprint(conn, "235 authenticated\r\n")
					case strings.HasPrefix(line, "MAIL FROM"), strings.HasPrefix(line, "RCPT TO"):
						if testCase.checkOnly {
							done <- fmt.Errorf("check-only operation tried to deliver email")
							return
						}
						_, _ = fmt.Fprint(conn, "250 accepted\r\n")
					case strings.HasPrefix(line, "DATA"):
						if !secured {
							done <- fmt.Errorf("message was sent without TLS")
							return
						}
						_, _ = fmt.Fprint(conn, "354 send message\r\n")
						data, err := io.ReadAll(textproto.NewReader(reader).DotReader())
						if err != nil {
							done <- err
							return
						}
						message = string(data)
						_, _ = fmt.Fprint(conn, "250 delivered\r\n")
					case strings.HasPrefix(line, "QUIT"):
						_, _ = fmt.Fprint(conn, "221 goodbye\r\n")
						if !testCase.checkOnly && !strings.Contains(message, "123456") {
							done <- fmt.Errorf("verification code missing")
							return
						}
						if !testCase.checkOnly {
							parsed, err := mail.ReadMessage(strings.NewReader(message))
							if err != nil || parsed.Header.Get("Date") == "" || parsed.Header.Get("Message-ID") == "" {
								done <- fmt.Errorf("required email headers missing")
								return
							}
						}
						done <- nil
						return
					default:
						done <- fmt.Errorf("unexpected SMTP command")
						return
					}
				}
			}()
			sender := SMTPSender{Config: SMTPConfig{Address: listener.Addr().String(), TLSMode: mode, Username: "noreply@klevliga.ru", Password: "test-only", From: "noreply@klevliga.ru"}, rootCAs: roots}
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			var operationErr error
			if testCase.checkOnly {
				operationErr = sender.Check(ctx)
			} else {
				operationErr = sender.SendVerificationCode(ctx, "recipient@example.test", "123456")
			}
			if operationErr != nil {
				t.Fatal(operationErr)
			}
			if err := <-done; err != nil {
				t.Fatal(err)
			}
		})
	}
}
