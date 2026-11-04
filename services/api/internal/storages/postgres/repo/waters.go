package repo

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/XATAB1CH/klev-liga/internal/domains/water"
	"github.com/XATAB1CH/klev-liga/internal/storages/postgres/sqlc"
)

type ListWatersFilter struct {
	RegionID   *uuid.UUID
	Q          *string
	CursorName *string
	CursorID   *uuid.UUID
	Limit      int32
}

func (s *Store) ListWaters(ctx context.Context, f ListWatersFilter) ([]*water.Water, error) {
	var regionID pgtype.UUID
	if f.RegionID != nil {
		regionID = uuidToPg(*f.RegionID)
	}

	var cursorID pgtype.UUID
	if f.CursorID != nil {
		cursorID = uuidToPg(*f.CursorID)
	}

	rows, err := s.queries(ctx).ListWaters(ctx, sqlc.ListWatersParams{
		RegionID:   regionID,
		Q:          f.Q,
		CursorName: f.CursorName,
		CursorID:   cursorID,
		Limit:      f.Limit,
	})
	if err != nil {
		return nil, wrapDBError("list waters", err)
	}

	waters := make([]*water.Water, 0, len(rows))
	for _, row := range rows {
		waters = append(waters, mapWater(row.ID, row.Slug, row.Name, row.Kind, row.RegionID, row.Description, row.CreatedAt))
	}

	return waters, nil
}

func (s *Store) GetWater(ctx context.Context, id uuid.UUID) (*water.Water, error) {
	row, err := s.queries(ctx).GetWater(ctx, id)
	if err != nil {
		return nil, wrapDBError("get water", err)
	}

	return mapWater(row.ID, row.Slug, row.Name, row.Kind, row.RegionID, row.Description, row.CreatedAt), nil
}

func (s *Store) SearchWaters(ctx context.Context, q string, limit int32) ([]*water.Water, error) {
	rows, err := s.queries(ctx).SearchWaters(ctx, sqlc.SearchWatersParams{
		Column1: &q,
		Limit:   limit,
	})
	if err != nil {
		return nil, wrapDBError("search waters", err)
	}

	waters := make([]*water.Water, 0, len(rows))
	for _, row := range rows {
		waters = append(waters, mapWater(row.ID, row.Slug, row.Name, row.Kind, row.RegionID, row.Description, row.CreatedAt))
	}

	return waters, nil
}
