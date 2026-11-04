package repo

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/XATAB1CH/klev-liga/internal/domains/forum"
	"github.com/XATAB1CH/klev-liga/internal/storages/postgres/sqlc"
)

func (s *Store) ListCategories(ctx context.Context) ([]*forum.Category, error) {
	rows, err := s.queries(ctx).ListCategories(ctx)
	if err != nil {
		return nil, wrapDBError("list categories", err)
	}

	categories := make([]*forum.Category, 0, len(rows))
	for _, row := range rows {
		categories = append(categories, mapCategory(row.ID, row.ParentID, row.Slug, row.Name, row.Position))
	}

	return categories, nil
}

func (s *Store) GetCategoryBySlug(ctx context.Context, slug string) (*forum.Category, error) {
	row, err := s.queries(ctx).GetCategoryBySlug(ctx, slug)
	if err != nil {
		return nil, wrapDBError("get category by slug", err)
	}

	return mapCategory(row.ID, row.ParentID, row.Slug, row.Name, row.Position), nil
}

func (s *Store) CreateTopic(ctx context.Context, categoryID, authorID uuid.UUID, slug, title, body, topicType string) (*forum.Topic, error) {
	row, err := s.queries(ctx).CreateTopic(ctx, sqlc.CreateTopicParams{
		CategoryID: categoryID,
		AuthorID:   authorID,
		Slug:       slug,
		Title:      title,
		Body:       body,
		Type:       topicType,
	})
	if err != nil {
		return nil, wrapDBError("create topic", err)
	}

	return mapTopic(row), nil
}

func (s *Store) GetTopic(ctx context.Context, id uuid.UUID) (*forum.Topic, error) {
	row, err := s.queries(ctx).GetTopic(ctx, id)
	if err != nil {
		return nil, wrapDBError("get topic", err)
	}

	return mapTopicRow(row.ID, row.CategoryID, row.AuthorID, row.Slug, row.Title, row.Body, row.Type,
		row.Pinned, row.Locked, row.RepliesCount, row.ViewsCount,
		row.LastActivityAt, row.CreatedAt, row.UpdatedAt, row.AuthorName, row.CategoryName), nil
}

type ListTopicsFilter struct {
	CategoryID *uuid.UUID
	CursorAt   *time.Time
	CursorID   *uuid.UUID
	Limit      int32
}

func (s *Store) ListTopics(ctx context.Context, f ListTopicsFilter) ([]*forum.Topic, error) {
	var categoryID pgtype.UUID
	if f.CategoryID != nil {
		categoryID = uuidToPg(*f.CategoryID)
	}

	var cursorAt pgtype.Timestamptz
	if f.CursorAt != nil {
		cursorAt = timeToPg(*f.CursorAt)
	}

	var cursorID pgtype.UUID
	if f.CursorID != nil {
		cursorID = uuidToPg(*f.CursorID)
	}

	rows, err := s.queries(ctx).ListTopics(ctx, sqlc.ListTopicsParams{
		CategoryID: categoryID,
		CursorAt:   cursorAt,
		CursorID:   cursorID,
		Limit:      f.Limit,
	})
	if err != nil {
		return nil, wrapDBError("list topics", err)
	}

	topics := make([]*forum.Topic, 0, len(rows))
	for _, row := range rows {
		topics = append(topics, mapTopicRow(row.ID, row.CategoryID, row.AuthorID, row.Slug, row.Title, row.Body, row.Type,
			row.Pinned, row.Locked, row.RepliesCount, row.ViewsCount,
			row.LastActivityAt, row.CreatedAt, row.UpdatedAt, row.AuthorName, row.CategoryName))
	}

	return topics, nil
}

type ListFeedFilter struct {
	Kind     *string
	CursorAt *time.Time
	CursorID *uuid.UUID
	Limit    int32
}

func (s *Store) ListFeed(ctx context.Context, f ListFeedFilter) ([]*forum.Topic, error) {
	var cursorAt pgtype.Timestamptz
	if f.CursorAt != nil {
		cursorAt = timeToPg(*f.CursorAt)
	}

	var cursorID pgtype.UUID
	if f.CursorID != nil {
		cursorID = uuidToPg(*f.CursorID)
	}

	rows, err := s.queries(ctx).ListFeed(ctx, sqlc.ListFeedParams{
		Kind:     f.Kind,
		CursorAt: cursorAt,
		CursorID: cursorID,
		Limit:    f.Limit,
	})
	if err != nil {
		return nil, wrapDBError("list feed", err)
	}

	topics := make([]*forum.Topic, 0, len(rows))
	for _, row := range rows {
		topics = append(topics, mapTopicRow(row.ID, row.CategoryID, row.AuthorID, row.Slug, row.Title, row.Body, row.Type,
			row.Pinned, row.Locked, row.RepliesCount, row.ViewsCount,
			row.LastActivityAt, row.CreatedAt, row.UpdatedAt, row.AuthorName, row.CategoryName))
	}

	return topics, nil
}

func (s *Store) SearchTopics(ctx context.Context, q string, limit int32) ([]*forum.Topic, error) {
	rows, err := s.queries(ctx).SearchTopics(ctx, sqlc.SearchTopicsParams{
		Q:     &q,
		Limit: limit,
	})
	if err != nil {
		return nil, wrapDBError("search topics", err)
	}

	topics := make([]*forum.Topic, 0, len(rows))
	for _, row := range rows {
		topics = append(topics, mapTopicRow(row.ID, row.CategoryID, row.AuthorID, row.Slug, row.Title, row.Body, row.Type,
			row.Pinned, row.Locked, row.RepliesCount, row.ViewsCount,
			row.LastActivityAt, row.CreatedAt, row.UpdatedAt, row.AuthorName, row.CategoryName))
	}

	return topics, nil
}

func (s *Store) CreatePost(ctx context.Context, topicID, authorID uuid.UUID, body string, parentPostID *uuid.UUID) (*forum.Post, error) {
	row, err := s.queries(ctx).CreatePost(ctx, sqlc.CreatePostParams{
		TopicID:      topicID,
		AuthorID:     authorID,
		Body:         body,
		ParentPostID: uuidPtrToPg(parentPostID),
	})
	if err != nil {
		return nil, wrapDBError("create post", err)
	}

	return mapPost(row), nil
}

type ListPostsFilter struct {
	TopicID  uuid.UUID
	CursorAt *time.Time
	CursorID *uuid.UUID
	Limit    int32
}

func (s *Store) ListPosts(ctx context.Context, f ListPostsFilter) ([]*forum.Post, error) {
	var cursorAt pgtype.Timestamptz
	if f.CursorAt != nil {
		cursorAt = timeToPg(*f.CursorAt)
	}

	var cursorID pgtype.UUID
	if f.CursorID != nil {
		cursorID = uuidToPg(*f.CursorID)
	}

	rows, err := s.queries(ctx).ListPosts(ctx, sqlc.ListPostsParams{
		TopicID:  f.TopicID,
		CursorAt: cursorAt,
		CursorID: cursorID,
		Limit:    f.Limit,
	})
	if err != nil {
		return nil, wrapDBError("list posts", err)
	}

	posts := make([]*forum.Post, 0, len(rows))
	for _, row := range rows {
		posts = append(posts, mapPostFields(row.ID, row.TopicID, row.AuthorID, row.Body, uuidPtr(row.ParentPostID), row.CreatedAt, timePtr(row.EditedAt), row.AuthorName))
	}

	return posts, nil
}

func (s *Store) TouchTopicActivity(ctx context.Context, topicID uuid.UUID) error {
	if err := s.queries(ctx).TouchTopicActivity(ctx, topicID); err != nil {
		return wrapDBError("touch topic activity", err)
	}

	return nil
}
