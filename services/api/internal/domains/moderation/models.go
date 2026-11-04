// Package moderation описывает жалобы на материалы (не путать с рыболовными отчётами report).
package moderation

import (
	"time"

	"github.com/google/uuid"
)

type AbuseReport struct {
	ID         uuid.UUID
	ReporterID uuid.UUID
	TargetType string
	TargetID   uuid.UUID
	Reason     string
	Status     string
	CreatedAt  time.Time
}

const (
	TargetTypeTopic  = "topic"
	TargetTypePost   = "post"
	TargetTypeReport = "report"
	TargetTypeUser   = "user"
)

const (
	StatusOpen     = "open"
	StatusResolved = "resolved"
	StatusRejected = "rejected"
)
