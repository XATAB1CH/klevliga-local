// Package report — бизнес-логика рыболовных отчётов.
package report

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/XATAB1CH/klev-liga/internal/domains/forum"
	"github.com/XATAB1CH/klev-liga/internal/domains/report"
	"github.com/XATAB1CH/klev-liga/internal/services/slug"
	"github.com/XATAB1CH/klev-liga/internal/storages/postgres/repo"
)

type Service struct {
	store *repo.Store
}

func NewService(store *repo.Store) *Service {
	return &Service{store: store}
}

type CreateParams struct {
	AuthorID           uuid.UUID
	WaterID            *uuid.UUID
	Title              string
	Body               string
	FishedAt           time.Time
	LocationVisibility string
	Species            []string
}

// reportCategorySlug — категория, в которую попадают темы-отчёты (см. seed в миграции).
const reportCategorySlug = "reports"

// Create создаёт отчёт как связанную тему + запись отчёта в одной транзакции
// (по модели данных: один отчёт = структурированные данные + одна тема форума).
func (s *Service) Create(ctx context.Context, p CreateParams) (*report.Report, error) {
	category, err := s.store.GetCategoryBySlug(ctx, reportCategorySlug)
	if err != nil {
		return nil, fmt.Errorf("resolve report category: %w", err)
	}

	var result *report.Report

	err = s.store.WithinTx(ctx, func(ctx context.Context) error {
		topic, err := s.store.CreateTopic(
			ctx,
			category.ID,
			p.AuthorID,
			slug.Unique(p.Title),
			p.Title,
			p.Body,
			forum.TopicTypeReport,
		)
		if err != nil {
			return err
		}

		result, err = s.store.CreateReport(ctx, repo.CreateReportParams{
			AuthorID:           p.AuthorID,
			TopicID:            topic.ID,
			WaterID:            p.WaterID,
			Title:              p.Title,
			Body:               p.Body,
			FishedAt:           p.FishedAt,
			LocationVisibility: p.LocationVisibility,
			Species:            p.Species,
		})
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	// Повторное чтение, чтобы вернуть отчёт с именем автора.
	return s.store.GetReport(ctx, result.ID)
}

func (s *Service) Get(ctx context.Context, id uuid.UUID) (*report.Report, error) {
	return s.store.GetReport(ctx, id)
}

func (s *Service) List(
	ctx context.Context,
	waterID *uuid.UUID,
	cursorAt *time.Time,
	cursorID *uuid.UUID,
	limit int32,
) ([]*report.Report, error) {
	return s.store.ListReports(ctx, repo.ListReportsFilter{
		WaterID:  waterID,
		CursorAt: cursorAt,
		CursorID: cursorID,
		Limit:    limit,
	})
}
