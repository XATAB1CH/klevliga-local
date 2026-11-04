package log

import (
	"log/slog"
	"os"
	"strings"

	"github.com/XATAB1CH/klev-liga/internal/config"
)

func New(cfg config.Logger) *slog.Logger {
	level := slog.LevelInfo

	switch strings.ToLower(cfg.Level) {
	case "debug":
		level = slog.LevelDebug
	case "warn", "warning":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	}

	opts := &slog.HandlerOptions{Level: level}

	var handler slog.Handler
	if cfg.Pretty {
		handler = slog.NewTextHandler(os.Stdout, opts)
	} else {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	}

	logger := slog.New(handler)
	if cfg.Title != "" {
		logger = logger.With("service", cfg.Title)
	}

	return logger
}
