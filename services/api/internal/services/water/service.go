// Package water — бизнес-логика каталога водоёмов.
package water

import (
	"context"

	"github.com/google/uuid"

	"github.com/XATAB1CH/klev-liga/internal/domains/water"
	"github.com/XATAB1CH/klev-liga/internal/storages/postgres/repo"
)

type Service struct {
	store *repo.Store
}

func NewService(store *repo.Store) *Service {
	return &Service{store: store}
}

func (s *Service) List(
	ctx context.Context,
	regionID *uuid.UUID,
	q *string,
	cursorName *string,
	cursorID *uuid.UUID,
	limit int32,
) ([]*water.Water, error) {
	return s.store.ListWaters(ctx, repo.ListWatersFilter{
		RegionID:   regionID,
		Q:          q,
		CursorName: cursorName,
		CursorID:   cursorID,
		Limit:      limit,
	})
}

func (s *Service) Get(ctx context.Context, id uuid.UUID) (*water.Water, error) {
	return s.store.GetWater(ctx, id)
}
