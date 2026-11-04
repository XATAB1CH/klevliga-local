package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"

	"github.com/XATAB1CH/klev-liga/internal/app"
	"github.com/XATAB1CH/klev-liga/internal/config"
	applog "github.com/XATAB1CH/klev-liga/pkg/log"
)

var (
	version   string
	buildTime string
)

func main() {
	log := applog.New(config.Logger{Title: "klevliga-api", Level: "info", Pretty: true})
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	defer cancel()

	log.Info("started", "version", version, "build_time", buildTime)

	defer func() {
		if msg := recover(); msg != nil {
			log.Error("recovered from panic, application will be terminated",
				"msg", fmt.Sprint(msg),
				"stack", string(debug.Stack()),
			)
			cancel()
		}
	}()

	var configPath string
	flag.StringVar(&configPath, "c", "./config.yaml", "path to config file")
	flag.Parse()

	cfg, err := config.New(ctx, configPath)
	if err != nil {
		log.Error("could not load config file", "error", err, "path", configPath)
		os.Exit(1)
	}

	log = applog.New(cfg.Logger)
	slog.SetDefault(log)
	log.Info("initialized logger from config")

	a, err := app.New(ctx, cfg, log)
	if err != nil {
		log.Error("could not initialize app", "error", err)
		os.Exit(1)
	}

	if err := a.Run(ctx); err != nil {
		log.Error("could not start application", "error", err)
		os.Exit(1)
	}
}
