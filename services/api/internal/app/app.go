// Package app — управление жизненным циклом приложения.
package app

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/XATAB1CH/klev-liga/internal/api/transport/http"
	"github.com/XATAB1CH/klev-liga/internal/app/provider"
	"github.com/XATAB1CH/klev-liga/internal/config"
	"github.com/XATAB1CH/klev-liga/migrations"
	"github.com/XATAB1CH/klev-liga/pkg/pg"
)

type App struct {
	cfg        *config.Config
	log        *slog.Logger
	provider   *provider.Provider
	httpServer *http.Server
}

func New(ctx context.Context, cfg *config.Config, log *slog.Logger) (*App, error) {
	db, err := pg.Connect(ctx, cfg.Connections.Postgres.Main)
	if err != nil {
		return nil, fmt.Errorf("connect postgres: %w", err)
	}

	if err := pg.MigrateFromPool(ctx, db, migrations.Migrations); err != nil {
		db.Close()

		return nil, fmt.Errorf("run migrations: %w", err)
	}

	p := provider.NewProvider(cfg, db)

	h := http.NewHandler(
		p.AuthService(),
		p.ForumService(),
		p.ReportService(),
		p.WaterService(),
		p.FeedService(),
		p.SearchService(),
		p.SocialService(),
		p.ModerationService(),
		cfg.Auth.CookieName,
		cfg.Auth.CookieSecure,
		cfg.Auth.CookieDomain,
		cfg.Auth.SessionTTL,
	)

	return &App{
		cfg:        cfg,
		log:        log,
		provider:   p,
		httpServer: http.NewServer(h, cfg.Server.HTTP.Address, log),
	}, nil
}

func (a *App) Run(ctx context.Context) error {
	defer a.Shutdown(context.Background())

	return a.httpServer.Start(ctx)
}

func (a *App) Shutdown(_ context.Context) {
	a.provider.Close()
}
