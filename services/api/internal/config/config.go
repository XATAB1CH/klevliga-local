package config

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Server      Server      `yaml:"server"`
	Connections Connections `yaml:"connections"`
	Auth        Auth        `yaml:"auth"`
	Mail        Mail        `yaml:"mail"`
	Logger      Logger      `yaml:"logger"`
}

type Server struct {
	HTTP BasicServer `yaml:"http"`
}

type BasicServer struct {
	Address            string `yaml:"address"`
	TrustLoopbackProxy bool   `yaml:"trust_loopback_proxy"`
}

type Connections struct {
	Postgres PostgresConnections `yaml:"postgres"`
}

type PostgresConnections struct {
	Main map[string]string `yaml:"main"`
}

type Auth struct {
	SessionTTL         time.Duration `yaml:"session_ttl"`
	CookieName         string        `yaml:"cookie_name"`
	CookieSecure       bool          `yaml:"cookie_secure"`
	CookieDomain       string        `yaml:"cookie_domain"`
	VerificationTTL    time.Duration `yaml:"verification_ttl"`
	VerificationSecret string        `yaml:"verification_secret"`
}

type Mail struct {
	Mode     string `yaml:"mode"`
	TLSMode  string `yaml:"tls_mode"`
	Address  string `yaml:"address"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	From     string `yaml:"from"`
}

type Logger struct {
	Title  string `yaml:"title"`
	Level  string `yaml:"level"`
	Pretty bool   `yaml:"pretty"`
}

var envPattern = regexp.MustCompile(`\$\{([A-Za-z_][A-Za-z0-9_]*)(?::([^}]*))?\}`)

func New(_ context.Context, path string) (*Config, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}
	fileEnv, err := godotenv.Read(filepath.Join(filepath.Dir(path), ".env"))
	if err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("read .env beside config: %w", err)
	}

	expanded := envPattern.ReplaceAllStringFunc(string(raw), func(match string) string {
		parts := envPattern.FindStringSubmatch(match)
		if value, ok := os.LookupEnv(parts[1]); ok {
			return value
		}
		if value, ok := fileEnv[parts[1]]; ok {
			return value
		}

		if len(parts) > 2 {
			return parts[2]
		}

		return ""
	})

	var cfg Config
	if err := yaml.Unmarshal([]byte(expanded), &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	if cfg.Server.HTTP.Address == "" {
		cfg.Server.HTTP.Address = ":8080"
	}

	if cfg.Auth.SessionTTL == 0 {
		cfg.Auth.SessionTTL = 30 * 24 * time.Hour
	}

	if cfg.Auth.CookieName == "" {
		cfg.Auth.CookieName = "klevliga_session"
	}
	if cfg.Auth.VerificationTTL == 0 {
		cfg.Auth.VerificationTTL = 10 * time.Minute
	}
	if cfg.Auth.VerificationSecret == "" {
		return nil, fmt.Errorf("auth.verification_secret is required")
	}
	if cfg.Mail.Mode == "" {
		cfg.Mail.Mode = "log"
	}
	if cfg.Mail.TLSMode == "" {
		cfg.Mail.TLSMode = "auto"
	}
	if cfg.Mail.TLSMode != "auto" && cfg.Mail.TLSMode != "tls" && cfg.Mail.TLSMode != "starttls" {
		return nil, fmt.Errorf("mail.tls_mode must be auto, tls or starttls")
	}
	if cfg.Mail.Mode != "log" && cfg.Mail.Mode != "smtp" {
		return nil, fmt.Errorf("mail.mode must be log or smtp")
	}
	if cfg.Mail.Mode == "smtp" && (cfg.Mail.Address == "" || cfg.Mail.From == "") {
		return nil, fmt.Errorf("mail.address and mail.from are required in smtp mode")
	}

	if cfg.Logger.Level == "" {
		cfg.Logger.Level = "info"
	}

	if cfg.Logger.Title == "" {
		cfg.Logger.Title = "klevliga-api"
	}

	return &cfg, nil
}
