package repo

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/XATAB1CH/klev-liga/internal/domains/report"
	"github.com/XATAB1CH/klev-liga/internal/storages/postgres/sqlc"
)

type CreateReportParams struct {
	AuthorID           uuid.UUID
	TopicID            uuid.UUID
	WaterID            *uuid.UUID
	Title              string
	Body               string
	FishedAt           time.Time
	LocationVisibility string
	Species            []string
}

func (s *Store) CreateReport(ctx context.Context, p CreateReportParams) (*report.Report, error) {
	row, err := s.queries(ctx).CreateReport(ctx, sqlc.CreateReportParams{
		AuthorID:           p.AuthorID,
		TopicID:            p.TopicID,
		WaterID:            uuidPtrToPg(p.WaterID),
		Title:              p.Title,
		Body:               p.Body,
		FishedAt:           p.FishedAt,
		LocationVisibility: p.LocationVisibility,
		Species:            p.Species,
	})
	if err != nil {
		return nil, wrapDBError("create report", err)
	}

	return mapReport(row), nil
}

func (s *Store) GetReport(ctx context.Context, id uuid.UUID) (*report.Report, error) {
	row, err := s.queries(ctx).GetReport(ctx, id)
	if err != nil {
		return nil, wrapDBError("get report", err)
	}

	return mapReportFields(row.ID, row.AuthorID, row.TopicID, uuidPtr(row.WaterID), row.Title, row.Body, row.FishedAt, row.LocationVisibility, row.Species, row.CreatedAt, row.AuthorName), nil
}

func (s *Store) GetReportByTopic(ctx context.Context, topicID uuid.UUID) (*report.Report, error) {
	row, err := s.queries(ctx).GetReportByTopic(ctx, topicID)
	if err != nil {
		return nil, wrapDBError("get report by topic", err)
	}

	return mapReportFields(row.ID, row.AuthorID, row.TopicID, uuidPtr(row.WaterID), row.Title, row.Body, row.FishedAt, row.LocationVisibility, row.Species, row.CreatedAt, row.AuthorName), nil
}

type ListReportsFilter struct {
	WaterID  *uuid.UUID
	CursorAt *time.Time
	CursorID *uuid.UUID
	Limit    int32
}

func (s *Store) ListReports(ctx context.Context, f ListReportsFilter) ([]*report.Report, error) {
	var waterID pgtype.UUID
	if f.WaterID != nil {
		waterID = uuidToPg(*f.WaterID)
	}

	var cursorAt pgtype.Timestamptz
	if f.CursorAt != nil {
		cursorAt = timeToPg(*f.CursorAt)
	}

	var cursorID pgtype.UUID
	if f.CursorID != nil {
		cursorID = uuidToPg(*f.CursorID)
	}

	rows, err := s.queries(ctx).ListReports(ctx, sqlc.ListReportsParams{
		WaterID:  waterID,
		CursorAt: cursorAt,
		CursorID: cursorID,
		Limit:    f.Limit,
	})
	if err != nil {
		return nil, wrapDBError("list reports", err)
	}

	reports := make([]*report.Report, 0, len(rows))
	for _, row := range rows {
		reports = append(reports, mapReportFields(row.ID, row.AuthorID, row.TopicID, uuidPtr(row.WaterID), row.Title, row.Body, row.FishedAt, row.LocationVisibility, row.Species, row.CreatedAt, row.AuthorName))
	}

	return reports, nil
}

func (s *Store) SearchReports(ctx context.Context, q string, limit int32) ([]*report.Report, error) {
	rows, err := s.queries(ctx).SearchReports(ctx, sqlc.SearchReportsParams{
		Column1: &q,
		Limit:   limit,
	})
	if err != nil {
		return nil, wrapDBError("search reports", err)
	}

	reports := make([]*report.Report, 0, len(rows))
	for _, row := range rows {
		reports = append(reports, mapReportFields(row.ID, row.AuthorID, row.TopicID, uuidPtr(row.WaterID), row.Title, row.Body, row.FishedAt, row.LocationVisibility, row.Species, row.CreatedAt, row.AuthorName))
	}

	return reports, nil
}
