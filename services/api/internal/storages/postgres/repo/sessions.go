package repo

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/XATAB1CH/klev-liga/internal/domains/auth"
	"github.com/XATAB1CH/klev-liga/internal/storages/postgres/sqlc"
)

func (s *Store) CreateSession(ctx context.Context, userID uuid.UUID, tokenHash []byte, expiresAt time.Time) (*auth.Session, error) {
	row, err := s.queries(ctx).CreateSession(ctx, sqlc.CreateSessionParams{
		UserID:    userID,
		TokenHash: tokenHash,
		ExpiresAt: expiresAt,
	})
	if err != nil {
		return nil, wrapDBError("create session", err)
	}

	return &auth.Session{
		ID:        row.ID,
		UserID:    row.UserID,
		TokenHash: tokenHash,
		ExpiresAt: row.ExpiresAt,
		CreatedAt: row.CreatedAt,
	}, nil
}

func (s *Store) GetSessionByTokenHash(ctx context.Context, tokenHash []byte) (*auth.Session, error) {
	row, err := s.queries(ctx).GetSessionByTokenHash(ctx, tokenHash)
	if err != nil {
		return nil, wrapDBError("get session by token hash", err)
	}

	return &auth.Session{
		ID:        row.ID,
		UserID:    row.UserID,
		TokenHash: tokenHash,
		ExpiresAt: row.ExpiresAt,
		RevokedAt: timePtr(row.RevokedAt),
	}, nil
}

func (s *Store) RevokeSession(ctx context.Context, tokenHash []byte) error {
	if err := s.queries(ctx).RevokeSession(ctx, tokenHash); err != nil {
		return wrapDBError("revoke session", err)
	}

	return nil
}
