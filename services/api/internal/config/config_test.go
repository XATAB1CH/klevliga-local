package config

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestEnvBesideConfigAndProcessPrecedence(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	configText := "auth:\n  verification_secret: test-secret\nmail:\n  mode: smtp\n  address: ${KLEVLIGA_TEST_SMTP_HOST}\n  password: ${KLEVLIGA_TEST_SMTP_PASS}\n  from: noreply@klevliga.ru\n"
	if err := os.WriteFile(configPath, []byte(configText), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".env"), []byte("KLEVLIGA_TEST_SMTP_HOST=mail.hosting.reg.ru:465\nKLEVLIGA_TEST_SMTP_PASS='demo#password'\n"), 0600); err != nil {
		t.Fatal(err)
	}
	cfg, err := New(context.Background(), configPath)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Mail.Address != "mail.hosting.reg.ru:465" || cfg.Mail.Password != "demo#password" || cfg.Mail.TLSMode != "auto" {
		t.Fatal(".env values or TLS default were not loaded")
	}
	if _, ok := os.LookupEnv("KLEVLIGA_TEST_SMTP_PASS"); ok {
		t.Fatal("config mutated the process environment")
	}
	t.Setenv("KLEVLIGA_TEST_SMTP_HOST", "override.example.test:587")
	cfg, err = New(context.Background(), configPath)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Mail.Address != "override.example.test:587" {
		t.Fatal("process environment must override .env")
	}
}

func TestEnvFileIsOptionalAndInvalidTLSModeFails(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	for _, tc := range []struct {
		mode  string
		valid bool
	}{{"auto", true}, {"tls", true}, {"starttls", true}, {"plain", false}} {
		if err := os.WriteFile(configPath, []byte("auth:\n  verification_secret: test-secret\nmail:\n  tls_mode: "+tc.mode+"\n"), 0600); err != nil {
			t.Fatal(err)
		}
		_, err := New(context.Background(), configPath)
		if (err == nil) != tc.valid {
			t.Errorf("TLS mode %s: error=%v", tc.mode, err)
		}
	}
}
