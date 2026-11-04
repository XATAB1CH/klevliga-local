package config

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server      Server      `yaml:"server"`
	Connections Connections `yaml:"connections"`
	Auth        Auth        `yaml:"auth"`
	Logger      Logger      `yaml:"logger"`
}

type Server struct {
	HTTP BasicServer `yaml:"http"`
}

type BasicServer struct {
	Address string `yaml:"address"`
}

type Connections struct {
	Postgres PostgresConnections `yaml:"postgres"`
}

type PostgresConnections struct {
	Main map[string]string `yaml:"main"`
}

type Auth struct {
	SessionTTL   time.Duration `yaml:"session_ttl"`
	CookieName   string        `yaml:"cookie_name"`
	CookieSecure bool          `yaml:"cookie_secure"`
	CookieDomain string        `yaml:"cookie_domain"`
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

	expanded := envPattern.ReplaceAllStringFunc(string(raw), func(match string) string {
		parts := envPattern.FindStringSubmatch(match)
		if value, ok := os.LookupEnv(parts[1]); ok {
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

	if cfg.Logger.Level == "" {
		cfg.Logger.Level = "info"
	}

	if cfg.Logger.Title == "" {
		cfg.Logger.Title = "klevliga-api"
	}

	return &cfg, nil
}
