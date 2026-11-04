// Package report описывает рыболовные отчёты (не путать с жалобами moderation).
package report

import (
	"time"

	"github.com/google/uuid"
)

// Report — один отчёт = структурированные данные + одна связанная тема форума.
type Report struct {
	ID                 uuid.UUID
	AuthorID           uuid.UUID
	TopicID            uuid.UUID
	WaterID            *uuid.UUID
	Title              string
	Body               string
	FishedAt           time.Time // дата (без времени)
	LocationVisibility string
	Species            []string
	CreatedAt          time.Time
	AuthorName         string
}

const (
	LocationVisibilityWaterOnly = "water_only"
	LocationVisibilityPrivate   = "private"
)
