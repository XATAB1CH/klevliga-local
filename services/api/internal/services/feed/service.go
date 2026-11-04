// Package feed — агрегированная публичная лента тем и отчётов.
package feed

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/XATAB1CH/klev-liga/internal/domains/forum"
	"github.com/XATAB1CH/klev-liga/internal/domains/report"
	"github.com/XATAB1CH/klev-liga/internal/storages/postgres/repo"
)

type Item struct {
	Topic  *forum.Topic
	Report *report.Report
}

type Service struct {
	store *repo.Store
}

func NewService(store *repo.Store) *Service {
	return &Service{store: store}
}

// Get возвращает ленту тем; для тем-отчётов дополнительно подгружается отчёт.
func (s *Service) Get(
	ctx context.Context,
	kind *string,
	cursorAt *time.Time,
	cursorID *uuid.UUID,
	limit int32,
) ([]*Item, error) {
	topics, err := s.store.ListFeed(ctx, repo.ListFeedFilter{
		Kind:     kind,
		CursorAt: cursorAt,
		CursorID: cursorID,
		Limit:    limit,
	})
	if err != nil {
		return nil, err
	}

	items := make([]*Item, 0, len(topics))
	for _, topic := range topics {
		item := &Item{Topic: topic}

		if topic.Type == forum.TopicTypeReport {
			rep, err := s.store.GetReportByTopic(ctx, topic.ID)
			if err != nil {
				return nil, err
			}

			item.Report = rep
		}

		items = append(items, item)
	}

	return items, nil
}
