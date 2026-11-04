// Package auth описывает пользователей и серверные сессии.
package auth

import (
	"time"

	"github.com/google/uuid"
)

// User — полная (внутренняя) модель пользователя, включая хэш пароля.
// Для публичных ответов используется отдельная DTO-структура без PasswordHash.
type User struct {
	ID           uuid.UUID
	Email        string
	Username     string
	PasswordHash string
	Status       string
	DisplayName  string
	Bio          *string
	Region       *string
	AvatarURL    *string
	Role         string
	CreatedAt    time.Time
}

// Session — серверная сессия. В БД хранится только хэш токена.
type Session struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	TokenHash []byte
	ExpiresAt time.Time
	RevokedAt *time.Time
	CreatedAt time.Time
}

const (
	UserStatusActive  = "active"
	UserStatusBlocked = "blocked"
)
