package repo

import (
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/XATAB1CH/klev-liga/internal/domains/auth"
	"github.com/XATAB1CH/klev-liga/internal/domains/forum"
	"github.com/XATAB1CH/klev-liga/internal/domains/moderation"
	"github.com/XATAB1CH/klev-liga/internal/domains/report"
	"github.com/XATAB1CH/klev-liga/internal/domains/social"
	"github.com/XATAB1CH/klev-liga/internal/domains/water"
	"github.com/XATAB1CH/klev-liga/internal/storages/postgres/sqlc"
)

func uuidPtr(v pgtype.UUID) *uuid.UUID {
	if !v.Valid {
		return nil
	}

	u := uuid.UUID(v.Bytes)

	return &u
}

func uuidToPg(id uuid.UUID) pgtype.UUID {
	return pgtype.UUID{Bytes: [16]byte(id), Valid: true}
}

func uuidPtrToPg(id *uuid.UUID) pgtype.UUID {
	if id == nil {
		return pgtype.UUID{}
	}

	return pgtype.UUID{Bytes: [16]byte(*id), Valid: true}
}

func timePtr(v pgtype.Timestamptz) *time.Time {
	if !v.Valid {
		return nil
	}

	t := v.Time

	return &t
}

func timeToPg(t time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: t, Valid: true}
}

func mapUser(
	id uuid.UUID,
	email, username, passwordHash, status string,
	createdAt time.Time,
	displayName string,
	bio, region, avatarURL *string,
	roles []string,
) *auth.User {
	return &auth.User{
		ID:           id,
		Email:        email,
		Username:     username,
		PasswordHash: passwordHash,
		Status:       status,
		DisplayName:  displayName,
		Bio:          bio,
		Region:       region,
		AvatarURL:    avatarURL,
		Roles:        roles,
		CreatedAt:    createdAt,
	}
}

func mapCategory(
	id uuid.UUID,
	parentID pgtype.UUID,
	slug, name string,
	position int32,
) *forum.Category {
	return &forum.Category{
		ID:       id,
		ParentID: uuidPtr(parentID),
		Slug:     slug,
		Name:     name,
		Position: position,
	}
}

func mapTopic(t sqlc.Topic) *forum.Topic {
	return mapTopicRow(
		t.ID, t.CategoryID, t.AuthorID, t.Slug, t.Title, t.Body, t.Type,
		t.Pinned, t.Locked, t.RepliesCount, t.ViewsCount,
		t.LastActivityAt, t.CreatedAt, t.UpdatedAt, "", "",
	)
}

func mapTopicRow(
	id, categoryID, authorID uuid.UUID,
	slug, title, body, topicType string,
	pinned, locked bool,
	repliesCount, viewsCount int32,
	lastActivityAt, createdAt, updatedAt time.Time,
	authorName, categoryName string,
) *forum.Topic {
	return &forum.Topic{
		ID:             id,
		CategoryID:     categoryID,
		AuthorID:       authorID,
		Slug:           slug,
		Title:          title,
		Body:           body,
		Type:           topicType,
		Pinned:         pinned,
		Locked:         locked,
		RepliesCount:   repliesCount,
		ViewsCount:     viewsCount,
		LastActivityAt: lastActivityAt,
		CreatedAt:      createdAt,
		UpdatedAt:      updatedAt,
		AuthorName:     authorName,
		CategoryName:   categoryName,
	}
}

func mapPost(p sqlc.TopicPost) *forum.Post {
	return mapPostFields(p.ID, p.TopicID, p.AuthorID, p.Body, uuidPtr(p.ParentPostID), p.CreatedAt, timePtr(p.EditedAt), "")
}

func mapPostFields(
	id, topicID, authorID uuid.UUID,
	body string,
	parentPostID *uuid.UUID,
	createdAt time.Time,
	editedAt *time.Time,
	authorName string,
) *forum.Post {
	return &forum.Post{
		ID:           id,
		TopicID:      topicID,
		AuthorID:     authorID,
		Body:         body,
		ParentPostID: parentPostID,
		CreatedAt:    createdAt,
		EditedAt:     editedAt,
		AuthorName:   authorName,
	}
}

func mapReport(r sqlc.Report) *report.Report {
	return mapReportFields(
		r.ID, r.AuthorID, r.TopicID, uuidPtr(r.WaterID),
		r.Title, r.Body, r.FishedAt, r.LocationVisibility, r.Species, r.CreatedAt, "",
	)
}

func mapReportFields(
	id, authorID, topicID uuid.UUID,
	waterID *uuid.UUID,
	title, body string,
	fishedAt time.Time,
	locationVisibility string,
	species []string,
	createdAt time.Time,
	authorName string,
) *report.Report {
	return &report.Report{
		ID:                 id,
		AuthorID:           authorID,
		TopicID:            topicID,
		WaterID:            waterID,
		Title:              title,
		Body:               body,
		FishedAt:           fishedAt,
		LocationVisibility: locationVisibility,
		Species:            species,
		CreatedAt:          createdAt,
		AuthorName:         authorName,
	}
}

func mapWater(
	id uuid.UUID,
	slug, name, kind string,
	regionID pgtype.UUID,
	description *string,
	createdAt time.Time,
) *water.Water {
	return &water.Water{
		ID:          id,
		Slug:        slug,
		Name:        name,
		Kind:        kind,
		RegionID:    uuidPtr(regionID),
		Description: description,
		CreatedAt:   createdAt,
	}
}

func mapNotification(n sqlc.Notification) *social.Notification {
	return &social.Notification{
		ID:          n.ID,
		RecipientID: n.RecipientID,
		Type:        n.Type,
		ActorID:     uuidPtr(n.ActorID),
		TargetType:  n.TargetType,
		TargetID:    uuidPtr(n.TargetID),
		Title:       n.Title,
		TargetURL:   n.TargetUrl,
		ReadAt:      timePtr(n.ReadAt),
		CreatedAt:   n.CreatedAt,
	}
}

func mapAbuseReport(
	id uuid.UUID,
	reporterID uuid.UUID,
	targetType string,
	targetID uuid.UUID,
	reason string,
	status string,
	createdAt time.Time,
) *moderation.AbuseReport {
	return &moderation.AbuseReport{
		ID:         id,
		ReporterID: reporterID,
		TargetType: targetType,
		TargetID:   targetID,
		Reason:     reason,
		Status:     status,
		CreatedAt:  createdAt,
	}
}
