package access

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// Changes contains only allowlisted service-produced fields, never request JSON.
type Event struct {
	ID         uuid.UUID       `json:"id"`
	ActorType  string          `json:"actor_type"`
	ActorID    *uuid.UUID      `json:"actor_id"`
	Action     string          `json:"action"`
	TargetType string          `json:"target_type"`
	TargetID   uuid.UUID       `json:"target_id"`
	Reason     string          `json:"reason"`
	RequestID  string          `json:"request_id"`
	Changes    json.RawMessage `json:"changes"`
	CreatedAt  time.Time       `json:"created_at"`
}

type AuditFilter struct {
	ActorID, TargetID *uuid.UUID
	Action            string
	From, Until       *time.Time
	BeforeTime        *time.Time
	BeforeID          *uuid.UUID
	Limit             int32
}
