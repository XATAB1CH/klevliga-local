// Package moderation — жалобы на материалы и модерация.
package moderation

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/XATAB1CH/klev-liga/internal/domains/apperr"
	"github.com/XATAB1CH/klev-liga/internal/domains/moderation"
	"github.com/XATAB1CH/klev-liga/internal/storages/postgres/repo"
)

type Service struct {
	store *repo.Store
}

func NewService(store *repo.Store) *Service {
	return &Service{store: store}
}

// Flag регистрирует жалобу на материал.
func (s *Service) Flag(ctx context.Context, reporterID uuid.UUID, targetType string, targetID uuid.UUID, reason string) (*moderation.AbuseReport, error) {
	if !validTargetType(targetType) || targetID == uuid.Nil {
		return nil, fmt.Errorf("%w: %s", apperr.ErrValidation, moderation.ErrCodeInvalidTargetType)
	}

	exists, err := s.store.ModerationTargetExists(ctx, targetType, targetID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, apperr.ErrNotFound
	}

	return s.store.CreateAbuseReport(ctx, reporterID, targetType, targetID, reason)
}

func validTargetType(t string) bool {
	switch t {
	case moderation.TargetTypeTopic,
		moderation.TargetTypePost,
		moderation.TargetTypeReport,
		moderation.TargetTypeUser:
		return true
	default:
		return false
	}
}
