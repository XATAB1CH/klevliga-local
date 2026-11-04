// Package auth — регистрация, вход, сессии и идентификация пользователя.
package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/XATAB1CH/klev-liga/internal/domains/apperr"
	authDomain "github.com/XATAB1CH/klev-liga/internal/domains/auth"
	"github.com/XATAB1CH/klev-liga/internal/storages/postgres/repo"
)

type Service struct {
	store      *repo.Store
	sessionTTL time.Duration
}

func NewService(store *repo.Store, sessionTTL time.Duration) *Service {
	return &Service{
		store:      store,
		sessionTTL: sessionTTL,
	}
}

// Register создаёт пользователя и профиль в одной транзакции.
// displayName по умолчанию равен username.
func (s *Service) Register(ctx context.Context, email, username, password string) error {
	hash, err := hashPassword(password)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	err = s.store.WithinTx(ctx, func(ctx context.Context) error {
		userID, err := s.store.CreateUser(ctx, email, username, hash)
		if err != nil {
			return err
		}

		if err := s.store.CreateProfile(ctx, userID, username); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

// Login проверяет учётные данные и создаёт сессию. Возвращает сырой токен
// (для установки cookie) и пользователя.
func (s *Service) Login(ctx context.Context, email, password string) (string, *authDomain.User, error) {
	user, err := s.store.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, apperr.ErrNotFound) {
			return "", nil, apperr.ErrInvalidCredentials
		}

		return "", nil, err
	}

	if user.Status != authDomain.UserStatusActive {
		return "", nil, fmt.Errorf("%w: account blocked", apperr.ErrForbidden)
	}

	ok, err := verifyPassword(password, user.PasswordHash)
	if err != nil {
		return "", nil, err
	}
	if !ok {
		return "", nil, apperr.ErrInvalidCredentials
	}

	raw, tokenHash, err := newSessionToken()
	if err != nil {
		return "", nil, fmt.Errorf("generate session token: %w", err)
	}

	if _, err := s.store.CreateSession(ctx, user.ID, tokenHash, time.Now().Add(s.sessionTTL)); err != nil {
		return "", nil, err
	}

	return raw, user, nil
}

// Logout отзывает сессию по сырому токену (идемпотентно).
func (s *Service) Logout(ctx context.Context, rawToken string) error {
	if rawToken == "" {
		return nil
	}

	if err := s.store.RevokeSession(ctx, hashToken(rawToken)); err != nil {
		if errors.Is(err, apperr.ErrNotFound) {
			return nil
		}

		return err
	}

	return nil
}

// Authenticate находит сессию по токену и возвращает пользователя.
func (s *Service) Authenticate(ctx context.Context, rawToken string) (*authDomain.User, error) {
	if rawToken == "" {
		return nil, apperr.ErrUnauthorized
	}

	session, err := s.store.GetSessionByTokenHash(ctx, hashToken(rawToken))
	if err != nil {
		if errors.Is(err, apperr.ErrNotFound) {
			return nil, apperr.ErrUnauthorized
		}

		return nil, err
	}

	if session.RevokedAt != nil || time.Now().After(session.ExpiresAt) {
		return nil, apperr.ErrUnauthorized
	}

	user, err := s.store.GetUserByID(ctx, session.UserID)
	if err != nil {
		return nil, err
	}

	if user.Status != authDomain.UserStatusActive {
		return nil, fmt.Errorf("%w: account blocked", apperr.ErrForbidden)
	}

	return user, nil
}

func (s *Service) GetUser(ctx context.Context, id uuid.UUID) (*authDomain.User, error) {
	return s.store.GetUserByID(ctx, id)
}
