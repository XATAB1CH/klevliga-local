// Package forum описывает категории, темы и ответы форума.
package forum

import (
	"time"

	"github.com/google/uuid"
)

type Category struct {
	ID       uuid.UUID
	ParentID *uuid.UUID
	Slug     string
	Name     string
	Position int32
}

type Topic struct {
	ID             uuid.UUID
	CategoryID     uuid.UUID
	AuthorID       uuid.UUID
	Slug           string
	Title          string
	Body           string
	Type           string
	Pinned         bool
	Locked         bool
	RepliesCount   int32
	ViewsCount     int32
	LastActivityAt time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time
	AuthorName     string
	CategoryName   string
}

type Post struct {
	ID           uuid.UUID
	TopicID      uuid.UUID
	AuthorID     uuid.UUID
	Body         string
	ParentPostID *uuid.UUID
	CreatedAt    time.Time
	EditedAt     *time.Time
	AuthorName   string
}

const (
	TopicTypeDiscussion = "discussion"
	TopicTypeQuestion   = "question"
	TopicTypeReport     = "report"
)
