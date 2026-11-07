package repo

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/XATAB1CH/klev-liga/internal/domains/auth"
	"github.com/XATAB1CH/klev-liga/internal/storages/postgres/sqlc"
)

type PendingRegistration struct {
	Email, Username, PasswordHash string
	CodeHash                      []byte
	Attempts                      int32
	ExpiresAt                     time.Time
}

func (s *Store) RegistrationAvailable(ctx context.Context, email, username string) (bool, error) {
	ok, err := s.queries(ctx).RegistrationAvailable(ctx, sqlc.RegistrationAvailableParams{Email: email, Username: username})
	if err != nil {
		return false, wrapDBError("check registration availability", err)
	}
	return ok, nil
}

func (s *Store) UpsertPendingRegistration(ctx context.Context, p PendingRegistration) error {
	err := s.queries(ctx).UpsertPendingRegistration(ctx, sqlc.UpsertPendingRegistrationParams{
		Email: p.Email, Username: p.Username, PasswordHash: p.PasswordHash,
		CodeHash: p.CodeHash, ExpiresAt: p.ExpiresAt,
	})
	if err != nil {
		return wrapDBError("save pending registration", err)
	}
	return nil
}

func (s *Store) GetPendingRegistrationForUpdate(ctx context.Context, email string) (*PendingRegistration, error) {
	r, err := s.queries(ctx).GetPendingRegistrationForUpdate(ctx, email)
	if err != nil {
		return nil, wrapDBError("get pending registration", err)
	}
	return &PendingRegistration{Email: r.Email, Username: r.Username, PasswordHash: r.PasswordHash, CodeHash: r.CodeHash, Attempts: r.Attempts, ExpiresAt: r.ExpiresAt}, nil
}

func (s *Store) IncrementPendingRegistrationAttempts(ctx context.Context, email string) error {
	if err := s.queries(ctx).IncrementPendingRegistrationAttempts(ctx, email); err != nil {
		return wrapDBError("increment verification attempts", err)
	}
	return nil
}

func (s *Store) DeletePendingRegistration(ctx context.Context, email string) error {
	if err := s.queries(ctx).DeletePendingRegistration(ctx, email); err != nil {
		return wrapDBError("delete pending registration", err)
	}
	return nil
}

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

func (s *Store) SetUserAvatar(ctx context.Context, userID uuid.UUID, imageData []byte, avatarURL string) error {
	if err := s.queries(ctx).UpsertUserAvatar(ctx, sqlc.UpsertUserAvatarParams{
		UserID: userID, ImageData: imageData,
	}); err != nil {
		return wrapDBError("save user avatar", err)
	}
	if err := s.queries(ctx).UpdateUserAvatarURL(ctx, sqlc.UpdateUserAvatarURLParams{
		UserID: userID, AvatarUrl: &avatarURL,
	}); err != nil {
		return wrapDBError("update user avatar url", err)
	}
	return nil
}

func (s *Store) GetUserAvatar(ctx context.Context, userID uuid.UUID) ([]byte, error) {
	imageData, err := s.queries(ctx).GetUserAvatar(ctx, userID)
	if err != nil {
		return nil, wrapDBError("get user avatar", err)
	}
	return imageData, nil
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
