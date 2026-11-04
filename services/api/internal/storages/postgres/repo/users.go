package repo

import (
	"context"

	"github.com/google/uuid"

	"github.com/XATAB1CH/klev-liga/internal/domains/auth"
	"github.com/XATAB1CH/klev-liga/internal/storages/postgres/sqlc"
)

// CreateUser создаёт пользователя и возвращает его ID.
func (s *Store) CreateUser(ctx context.Context, email, username, passwordHash string) (uuid.UUID, error) {
	row, err := s.queries(ctx).CreateUser(ctx, sqlc.CreateUserParams{
		Email:        email,
		Username:     username,
		PasswordHash: passwordHash,
	})
	if err != nil {
		return uuid.Nil, wrapDBError("create user", err)
	}

	return row.ID, nil
}

func (s *Store) CreateProfile(ctx context.Context, userID uuid.UUID, displayName string) error {
	_, err := s.queries(ctx).CreateProfile(ctx, sqlc.CreateProfileParams{
		UserID:      userID,
		DisplayName: displayName,
	})
	if err != nil {
		return wrapDBError("create profile", err)
	}

	return nil
}

func (s *Store) GetUserByEmail(ctx context.Context, email string) (*auth.User, error) {
	row, err := s.queries(ctx).GetUserByEmail(ctx, email)
	if err != nil {
		return nil, wrapDBError("get user by email", err)
	}

	return mapUser(
		row.ID, row.Email, row.Username, row.PasswordHash, row.Status, row.CreatedAt,
		row.DisplayName, row.Bio, row.Region, row.AvatarUrl, row.Role,
	), nil
}

func (s *Store) GetUserByID(ctx context.Context, id uuid.UUID) (*auth.User, error) {
	row, err := s.queries(ctx).GetUserByID(ctx, id)
	if err != nil {
		return nil, wrapDBError("get user by id", err)
	}

	return mapUser(
		row.ID, row.Email, row.Username, row.PasswordHash, row.Status, row.CreatedAt,
		row.DisplayName, row.Bio, row.Region, row.AvatarUrl, row.Role,
	), nil
}

func (s *Store) SearchUsers(ctx context.Context, q string, limit int32) ([]*auth.User, error) {
	rows, err := s.queries(ctx).SearchUsers(ctx, sqlc.SearchUsersParams{
		Column1: &q,
		Limit:   limit,
	})
	if err != nil {
		return nil, wrapDBError("search users", err)
	}

	users := make([]*auth.User, 0, len(rows))
	for _, row := range rows {
		users = append(users, mapUser(
			row.ID, "", row.Username, "", auth.UserStatusActive, row.CreatedAt,
			row.DisplayName, row.Bio, row.Region, row.AvatarUrl, "",
		))
	}

	return users, nil
}
