package repo

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/XATAB1CH/klev-liga/internal/domains/social"
	"github.com/XATAB1CH/klev-liga/internal/storages/postgres/sqlc"
)

func (s *Store) Follow(ctx context.Context, userID uuid.UUID, targetType string, targetID uuid.UUID) error {
	if err := s.queries(ctx).Follow(ctx, sqlc.FollowParams{
		UserID:     userID,
		TargetType: targetType,
		TargetID:   targetID,
	}); err != nil {
		return wrapDBError("follow", err)
	}

	return nil
}

func (s *Store) Unfollow(ctx context.Context, userID uuid.UUID, targetType string, targetID uuid.UUID) error {
	if err := s.queries(ctx).Unfollow(ctx, sqlc.UnfollowParams{
		UserID:     userID,
		TargetType: targetType,
		TargetID:   targetID,
	}); err != nil {
		return wrapDBError("unfollow", err)
	}

	return nil
}

func (s *Store) NotifyTopicSubscribers(ctx context.Context, topicID, actorID uuid.UUID) error {
	err := s.queries(ctx).NotifyTopicSubscribers(ctx, sqlc.NotifyTopicSubscribersParams{
		TopicID: topicID, ActorID: actorID, TargetUrl: "/forum/topics/" + topicID.String(),
	})
	return wrapDBError("notify topic subscribers", err)
}

func (s *Store) NotifyWaterSubscribers(ctx context.Context, waterID, topicID, actorID uuid.UUID) error {
	err := s.queries(ctx).NotifyWaterSubscribers(ctx, sqlc.NotifyWaterSubscribersParams{
		WaterID: waterID, TopicID: topicID, ActorID: actorID, TargetUrl: "/forum/topics/" + topicID.String(),
	})
	return wrapDBError("notify water subscribers", err)
}

type CreateNotificationParams struct {
	RecipientID uuid.UUID
	Type        string
	ActorID     *uuid.UUID
	TargetType  *string
	TargetID    *uuid.UUID
	Title       string
	TargetURL   *string
}

func (s *Store) CreateNotification(ctx context.Context, p CreateNotificationParams) (*social.Notification, error) {
	row, err := s.queries(ctx).CreateNotification(ctx, sqlc.CreateNotificationParams{
		RecipientID: p.RecipientID,
		Type:        p.Type,
		ActorID:     uuidPtrToPg(p.ActorID),
		TargetType:  p.TargetType,
		TargetID:    uuidPtrToPg(p.TargetID),
		Title:       p.Title,
		TargetUrl:   p.TargetURL,
	})
	if err != nil {
		return nil, wrapDBError("create notification", err)
	}

	return mapNotification(row), nil
}

type ListNotificationsFilter struct {
	RecipientID uuid.UUID
	CursorAt    *time.Time
	CursorID    *uuid.UUID
	Limit       int32
}

func (s *Store) ListNotifications(ctx context.Context, f ListNotificationsFilter) ([]*social.Notification, error) {
	var cursorAt pgtype.Timestamptz
	if f.CursorAt != nil {
		cursorAt = timeToPg(*f.CursorAt)
	}

	var cursorID pgtype.UUID
	if f.CursorID != nil {
		cursorID = uuidToPg(*f.CursorID)
	}

	rows, err := s.queries(ctx).ListNotifications(ctx, sqlc.ListNotificationsParams{
		RecipientID: f.RecipientID,
		CursorAt:    cursorAt,
		CursorID:    cursorID,
		Limit:       f.Limit,
	})
	if err != nil {
		return nil, wrapDBError("list notifications", err)
	}

	notifications := make([]*social.Notification, 0, len(rows))
	for _, row := range rows {
		notifications = append(notifications, mapNotification(row))
	}

	return notifications, nil
}
