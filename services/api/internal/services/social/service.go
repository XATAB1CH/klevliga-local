// Package social — подписки и уведомления.
package social

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/XATAB1CH/klev-liga/internal/domains/apperr"
	"github.com/XATAB1CH/klev-liga/internal/domains/social"
	"github.com/XATAB1CH/klev-liga/internal/storages/postgres/repo"
)

type Service struct {
	store *repo.Store
}

func NewService(store *repo.Store) *Service {
	return &Service{store: store}
}

func (s *Service) Follow(ctx context.Context, userID uuid.UUID, targetType string, targetID uuid.UUID) error {
	if !validTargetType(targetType) {
		return fmt.Errorf("%w: %s", apperr.ErrValidation, social.ErrCodeInvalidTargetType)
	}

	return s.store.Follow(ctx, userID, targetType, targetID)
}

func (s *Service) Unfollow(ctx context.Context, userID uuid.UUID, targetType string, targetID uuid.UUID) error {
	if !validTargetType(targetType) {
		return fmt.Errorf("%w: %s", apperr.ErrValidation, social.ErrCodeInvalidTargetType)
	}

	return s.store.Unfollow(ctx, userID, targetType, targetID)
}

func (s *Service) ListNotifications(
	ctx context.Context,
	recipientID uuid.UUID,
	cursorAt *time.Time,
	cursorID *uuid.UUID,
	limit int32,
) ([]*social.Notification, error) {
	return s.store.ListNotifications(ctx, repo.ListNotificationsFilter{
		RecipientID: recipientID,
		CursorAt:    cursorAt,
		CursorID:    cursorID,
		Limit:       limit,
	})
}

func validTargetType(t string) bool {
	return t == social.FollowTargetTopic || t == social.FollowTargetWater
}
