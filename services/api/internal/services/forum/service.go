// Package forum — бизнес-логика форума (категории, темы, ответы).
package forum

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/XATAB1CH/klev-liga/internal/domains/apperr"
	"github.com/XATAB1CH/klev-liga/internal/domains/forum"
	"github.com/XATAB1CH/klev-liga/internal/domains/social"
	"github.com/XATAB1CH/klev-liga/internal/services/slug"
	"github.com/XATAB1CH/klev-liga/internal/storages/postgres/repo"
)

type Service struct {
	store *repo.Store
}

func NewService(store *repo.Store) *Service {
	return &Service{store: store}
}

func (s *Service) ListCategories(ctx context.Context) ([]*forum.Category, error) {
	return s.store.ListCategories(ctx)
}

func (s *Service) CreateTopic(ctx context.Context, categoryID, authorID uuid.UUID, title, body string) (*forum.Topic, error) {
	topic, err := s.store.CreateTopic(ctx, categoryID, authorID, slug.Unique(title), title, body, forum.TopicTypeDiscussion)
	if err != nil {
		return nil, err
	}

	// Повторное чтение, чтобы вернуть тему с именами автора и категории.
	return s.store.GetTopic(ctx, topic.ID)
}

func (s *Service) GetTopic(ctx context.Context, id uuid.UUID) (*forum.Topic, error) {
	return s.store.GetTopic(ctx, id)
}

func (s *Service) ListTopics(
	ctx context.Context,
	categoryID *uuid.UUID,
	cursorAt *time.Time,
	cursorID *uuid.UUID,
	limit int32,
) ([]*forum.Topic, error) {
	return s.store.ListTopics(ctx, repo.ListTopicsFilter{
		CategoryID: categoryID,
		CursorAt:   cursorAt,
		CursorID:   cursorID,
		Limit:      limit,
	})
}

// CreatePost создаёт ответ, обновляет активность темы и уведомляет автора темы.
func (s *Service) CreatePost(
	ctx context.Context,
	topicID, authorID uuid.UUID,
	body string,
	parentPostID *uuid.UUID,
) (*forum.Post, error) {
	topic, err := s.store.GetTopic(ctx, topicID)
	if err != nil {
		return nil, err
	}

	if topic.Locked {
		return nil, fmt.Errorf("%w: topic is locked", apperr.ErrForbidden)
	}

	var post *forum.Post

	err = s.store.WithinTx(ctx, func(ctx context.Context) error {
		post, err = s.store.CreatePost(ctx, topicID, authorID, body, parentPostID)
		if err != nil {
			return err
		}

		if err := s.store.TouchTopicActivity(ctx, topicID); err != nil {
			return err
		}

		if topic.AuthorID != authorID {
			targetType := social.FollowTargetTopic
			if _, err := s.store.CreateNotification(ctx, repo.CreateNotificationParams{
				RecipientID: topic.AuthorID,
				Type:        social.NotificationTopicReply,
				ActorID:     &authorID,
				TargetType:  &targetType,
				TargetID:    &topicID,
				Title:       "Новый ответ в вашей теме",
			}); err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return post, nil
}

func (s *Service) ListPosts(
	ctx context.Context,
	topicID uuid.UUID,
	cursorAt *time.Time,
	cursorID *uuid.UUID,
	limit int32,
) ([]*forum.Post, error) {
	return s.store.ListPosts(ctx, repo.ListPostsFilter{
		TopicID:  topicID,
		CursorAt: cursorAt,
		CursorID: cursorID,
		Limit:    limit,
	})
}
