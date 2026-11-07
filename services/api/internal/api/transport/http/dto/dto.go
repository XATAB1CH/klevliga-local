// Package dto — структуры запросов/ответов REST API и мапперы domain<->dto.
package dto

import (
	"time"

	"github.com/google/uuid"

	"github.com/XATAB1CH/klev-liga/internal/domains/auth"
	"github.com/XATAB1CH/klev-liga/internal/domains/forum"
	"github.com/XATAB1CH/klev-liga/internal/domains/report"
	"github.com/XATAB1CH/klev-liga/internal/domains/social"
	"github.com/XATAB1CH/klev-liga/internal/domains/water"
)

// --- Error ---

type ApiError struct {
	Error ErrorBody `json:"error"`
}

type ErrorBody struct {
	Code      string            `json:"code"`
	Message   string            `json:"message"`
	RequestID string            `json:"request_id"`
	Fields    map[string]string `json:"fields,omitempty"`
}

// --- Auth / Users ---

type RegisterRequest struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type VerifyRegistrationRequest struct {
	Email string `json:"email"`
	Code  string `json:"code"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// User — публичный профиль.
// NOTE: черновой OpenAPI описывает `region_id` как uuid, но в миграции
// user_profiles.region — это text; здесь регион отдаётся строкой.
type User struct {
	ID          uuid.UUID `json:"id"`
	Username    string    `json:"username"`
	DisplayName string    `json:"display_name"`
	AvatarURL   *string   `json:"avatar_url"`
	Region      *string   `json:"region"`
	Bio         *string   `json:"bio"`
	JoinedAt    time.Time `json:"joined_at"`
}

// SessionUser adds private authorization data only to /auth/me.
type SessionUser struct {
	*User
	Roles       []string          `json:"roles"`
	Permissions []auth.Permission `json:"permissions"`
}

func UserFromDomain(u *auth.User) *User {
	if u == nil {
		return nil
	}

	return &User{
		ID:          u.ID,
		Username:    u.Username,
		DisplayName: u.DisplayName,
		AvatarURL:   u.AvatarURL,
		Region:      u.Region,
		Bio:         u.Bio,
		JoinedAt:    u.CreatedAt,
	}
}

// --- Forum ---

type Category struct {
	ID       uuid.UUID  `json:"id"`
	Name     string     `json:"name"`
	Slug     string     `json:"slug"`
	ParentID *uuid.UUID `json:"parent_id"`
}

func CategoryFromDomain(c *forum.Category) *Category {
	if c == nil {
		return nil
	}

	return &Category{
		ID:       c.ID,
		Name:     c.Name,
		Slug:     c.Slug,
		ParentID: c.ParentID,
	}
}

type Topic struct {
	ID             uuid.UUID `json:"id"`
	CategoryID     uuid.UUID `json:"category_id"`
	AuthorID       uuid.UUID `json:"author_id"`
	Title          string    `json:"title"`
	Slug           string    `json:"slug"`
	Body           string    `json:"body"`
	Type           string    `json:"type"`
	Pinned         bool      `json:"pinned"`
	Locked         bool      `json:"locked"`
	CreatedAt      time.Time `json:"created_at"`
	LastActivityAt time.Time `json:"last_activity_at"`
	RepliesCount   int32     `json:"replies_count"`
	ViewsCount     int32     `json:"views_count"`
	AuthorName     string    `json:"author_name"`
	CategoryName   string    `json:"category_name"`
}

func TopicFromDomain(t *forum.Topic) *Topic {
	if t == nil {
		return nil
	}

	return &Topic{
		ID:             t.ID,
		CategoryID:     t.CategoryID,
		AuthorID:       t.AuthorID,
		Title:          t.Title,
		Slug:           t.Slug,
		Body:           t.Body,
		Type:           t.Type,
		Pinned:         t.Pinned,
		Locked:         t.Locked,
		CreatedAt:      t.CreatedAt,
		LastActivityAt: t.LastActivityAt,
		RepliesCount:   t.RepliesCount,
		ViewsCount:     t.ViewsCount,
		AuthorName:     t.AuthorName,
		CategoryName:   t.CategoryName,
	}
}

type NewTopicRequest struct {
	CategoryID uuid.UUID `json:"category_id"`
	Title      string    `json:"title"`
	Body       string    `json:"body"`
}

type Post struct {
	ID           uuid.UUID  `json:"id"`
	TopicID      uuid.UUID  `json:"topic_id"`
	AuthorID     uuid.UUID  `json:"author_id"`
	Body         string     `json:"body"`
	CreatedAt    time.Time  `json:"created_at"`
	EditedAt     *time.Time `json:"edited_at"`
	ParentPostID *uuid.UUID `json:"parent_post_id"`
	AuthorName   string     `json:"author_name"`
}

func PostFromDomain(p *forum.Post) *Post {
	if p == nil {
		return nil
	}

	return &Post{
		ID:           p.ID,
		TopicID:      p.TopicID,
		AuthorID:     p.AuthorID,
		Body:         p.Body,
		CreatedAt:    p.CreatedAt,
		EditedAt:     p.EditedAt,
		ParentPostID: p.ParentPostID,
		AuthorName:   p.AuthorName,
	}
}

type ReplyRequest struct {
	Body         string     `json:"body"`
	ParentPostID *uuid.UUID `json:"parent_post_id"`
}

// --- Reports ---

type Report struct {
	ID                 uuid.UUID  `json:"id"`
	TopicID            uuid.UUID  `json:"topic_id"`
	AuthorID           uuid.UUID  `json:"author_id"`
	Title              string     `json:"title"`
	Body               string     `json:"body"`
	WaterID            *uuid.UUID `json:"water_id"`
	FishedAt           string     `json:"fished_at"`
	LocationVisibility string     `json:"location_visibility"`
	Species            []string   `json:"species"`
	MediaURLs          []string   `json:"media_urls"`
	CreatedAt          time.Time  `json:"created_at"`
	AuthorName         string     `json:"author_name"`
}

func ReportFromDomain(r *report.Report) *Report {
	if r == nil {
		return nil
	}

	return &Report{
		ID:                 r.ID,
		TopicID:            r.TopicID,
		AuthorID:           r.AuthorID,
		Title:              r.Title,
		Body:               r.Body,
		WaterID:            r.WaterID,
		FishedAt:           r.FishedAt.Format("2006-01-02"),
		LocationVisibility: r.LocationVisibility,
		Species:            r.Species,
		MediaURLs:          []string{},
		CreatedAt:          r.CreatedAt,
		AuthorName:         r.AuthorName,
	}
}

type NewReportRequest struct {
	Title              string      `json:"title"`
	Body               string      `json:"body"`
	WaterID            *uuid.UUID  `json:"water_id"`
	FishedAt           string      `json:"fished_at"`
	LocationVisibility string      `json:"location_visibility"`
	MediaIDs           []uuid.UUID `json:"media_ids"`
	Species            []string    `json:"species"`
}

// --- Waters ---

type Water struct {
	ID          uuid.UUID  `json:"id"`
	Name        string     `json:"name"`
	Slug        string     `json:"slug"`
	Kind        string     `json:"kind"`
	RegionID    *uuid.UUID `json:"region_id"`
	Description *string    `json:"description"`
	CoverURL    *string    `json:"cover_url"`
}

func WaterFromDomain(w *water.Water) *Water {
	if w == nil {
		return nil
	}

	return &Water{
		ID:          w.ID,
		Name:        w.Name,
		Slug:        w.Slug,
		Kind:        w.Kind,
		RegionID:    w.RegionID,
		Description: w.Description,
		CoverURL:    nil,
	}
}

// --- Feed ---

type FeedItem struct {
	Kind   string  `json:"kind"`
	Topic  *Topic  `json:"topic"`
	Report *Report `json:"report"`
}

// --- Notifications ---

type Notification struct {
	ID        uuid.UUID `json:"id"`
	Type      string    `json:"type"`
	Title     string    `json:"title"`
	CreatedAt time.Time `json:"created_at"`
	Read      bool      `json:"read"`
	TargetURL string    `json:"target_url"`
}

func NotificationFromDomain(n *social.Notification) *Notification {
	if n == nil {
		return nil
	}

	targetURL := ""
	if n.TargetURL != nil {
		targetURL = *n.TargetURL
	}

	return &Notification{
		ID:        n.ID,
		Type:      n.Type,
		Title:     n.Title,
		CreatedAt: n.CreatedAt,
		Read:      n.ReadAt != nil,
		TargetURL: targetURL,
	}
}

// --- Follows ---

type FollowRequest struct {
	TargetType string    `json:"target_type"`
	TargetID   uuid.UUID `json:"target_id"`
}

// --- Moderation ---

type FlagRequest struct {
	TargetType string    `json:"target_type"`
	TargetID   uuid.UUID `json:"target_id"`
	Reason     string    `json:"reason"`
}

// --- Pages ---

type TopicPage struct {
	Items      []*Topic `json:"items"`
	NextCursor *string  `json:"next_cursor"`
}

type PostPage struct {
	Items      []*Post `json:"items"`
	NextCursor *string `json:"next_cursor"`
}

type ReportPage struct {
	Items      []*Report `json:"items"`
	NextCursor *string   `json:"next_cursor"`
}

type WaterPage struct {
	Items      []*Water `json:"items"`
	NextCursor *string  `json:"next_cursor"`
}

type FeedPage struct {
	Items      []*FeedItem `json:"items"`
	NextCursor *string     `json:"next_cursor"`
}

type NotificationPage struct {
	Items      []*Notification `json:"items"`
	NextCursor *string         `json:"next_cursor"`
}
