// Package social описывает подписки и уведомления.
package social

import (
	"time"

	"github.com/google/uuid"
)

type Notification struct {
	ID          uuid.UUID
	RecipientID uuid.UUID
	Type        string
	ActorID     *uuid.UUID
	TargetType  *string
	TargetID    *uuid.UUID
	Title       string
	TargetURL   *string
	ReadAt      *time.Time
	CreatedAt   time.Time
}

const (
	FollowTargetTopic = "topic"
	FollowTargetWater = "water"
)

const (
	NotificationTopicReply    = "topic_reply"
	NotificationFollowedTopic = "followed_topic"
	NotificationFollowedWater = "followed_water"
	NotificationMention       = "mention"
	NotificationModeration    = "moderation"
)
