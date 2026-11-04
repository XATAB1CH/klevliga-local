// Package provider — ленивая композиция зависимостей (хранилища и сервисы).
package provider

import (
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/XATAB1CH/klev-liga/internal/config"
	"github.com/XATAB1CH/klev-liga/internal/services/auth"
	"github.com/XATAB1CH/klev-liga/internal/services/feed"
	"github.com/XATAB1CH/klev-liga/internal/services/forum"
	"github.com/XATAB1CH/klev-liga/internal/services/moderation"
	"github.com/XATAB1CH/klev-liga/internal/services/report"
	"github.com/XATAB1CH/klev-liga/internal/services/search"
	"github.com/XATAB1CH/klev-liga/internal/services/social"
	"github.com/XATAB1CH/klev-liga/internal/services/water"
	"github.com/XATAB1CH/klev-liga/internal/storages/postgres/repo"
)

type Provider struct {
	cfg *config.Config
	db  *pgxpool.Pool

	store     *repo.Store
	authSvc   *auth.Service
	forumSvc  *forum.Service
	reportSvc *report.Service
	waterSvc  *water.Service
	feedSvc   *feed.Service
	searchSvc *search.Service
	socialSvc *social.Service
	modSvc    *moderation.Service
}

func NewProvider(cfg *config.Config, db *pgxpool.Pool) *Provider {
	return &Provider{cfg: cfg, db: db}
}

func (p *Provider) Config() *config.Config {
	return p.cfg
}

func (p *Provider) Store() *repo.Store {
	if p.store == nil {
		p.store = repo.NewStore(p.db)
	}

	return p.store
}

func (p *Provider) AuthService() *auth.Service {
	if p.authSvc == nil {
		p.authSvc = auth.NewService(p.Store(), p.cfg.Auth.SessionTTL)
	}

	return p.authSvc
}

func (p *Provider) ForumService() *forum.Service {
	if p.forumSvc == nil {
		p.forumSvc = forum.NewService(p.Store())
	}

	return p.forumSvc
}

func (p *Provider) ReportService() *report.Service {
	if p.reportSvc == nil {
		p.reportSvc = report.NewService(p.Store())
	}

	return p.reportSvc
}

func (p *Provider) WaterService() *water.Service {
	if p.waterSvc == nil {
		p.waterSvc = water.NewService(p.Store())
	}

	return p.waterSvc
}

func (p *Provider) FeedService() *feed.Service {
	if p.feedSvc == nil {
		p.feedSvc = feed.NewService(p.Store())
	}

	return p.feedSvc
}

func (p *Provider) SearchService() *search.Service {
	if p.searchSvc == nil {
		p.searchSvc = search.NewService(p.Store())
	}

	return p.searchSvc
}

func (p *Provider) SocialService() *social.Service {
	if p.socialSvc == nil {
		p.socialSvc = social.NewService(p.Store())
	}

	return p.socialSvc
}

func (p *Provider) ModerationService() *moderation.Service {
	if p.modSvc == nil {
		p.modSvc = moderation.NewService(p.Store())
	}

	return p.modSvc
}

// Close закрывает пул соединений с БД.
func (p *Provider) Close() {
	if p.db != nil {
		p.db.Close()
	}
}
