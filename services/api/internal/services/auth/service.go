// Package auth — регистрация, вход, сессии и идентификация пользователя.
package auth

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/XATAB1CH/klev-liga/internal/domains/apperr"
	authDomain "github.com/XATAB1CH/klev-liga/internal/domains/auth"
	mailSvc "github.com/XATAB1CH/klev-liga/internal/services/mail"
	"github.com/XATAB1CH/klev-liga/internal/storages/postgres/repo"
)

type Service struct {
	store              *repo.Store
	sessionTTL         time.Duration
	verificationTTL    time.Duration
	verificationSecret []byte
	mail               mailSvc.Sender
}

func NewService(store *repo.Store, sessionTTL, verificationTTL time.Duration, verificationSecret string, sender mailSvc.Sender) *Service {
	return &Service{
		store:              store,
		sessionTTL:         sessionTTL,
		verificationTTL:    verificationTTL,
		verificationSecret: []byte(verificationSecret),
		mail:               sender,
	}
}

// Register создаёт временную заявку и отправляет код подтверждения.
func (s *Service) Register(ctx context.Context, email, username, password string) error {
	available, err := s.store.RegistrationAvailable(ctx, email, username)
	if err != nil {
		return err
	}
	if !available {
		return apperr.ErrConflict
	}
	hash, err := hashPassword(password)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	code, err := newVerificationCode()
	if err != nil {
		return fmt.Errorf("generate verification code: %w", err)
	}
	if err := s.store.UpsertPendingRegistration(ctx, repo.PendingRegistration{
		Email: email, Username: username, PasswordHash: hash,
		CodeHash: s.verificationHash(email, code), ExpiresAt: time.Now().Add(s.verificationTTL),
	}); err != nil {
		return err
	}
	if err := s.mail.SendVerificationCode(ctx, email, code); err != nil {
		_ = s.store.DeletePendingRegistration(ctx, email)
		return fmt.Errorf("send verification email: %w", err)
	}
	return nil
}

const maxVerificationAttempts = 5

func (s *Service) VerifyRegistration(ctx context.Context, email, code string) (string, *authDomain.User, error) {
	raw, tokenHash, err := newSessionToken()
	if err != nil {
		return "", nil, fmt.Errorf("generate session token: %w", err)
	}
	var userID uuid.UUID
	var verificationErr error
	err = s.store.WithinTx(ctx, func(ctx context.Context) error {
		pending, err := s.store.GetPendingRegistrationForUpdate(ctx, email)
		if err != nil {
			if errors.Is(err, apperr.ErrNotFound) {
				verificationErr = apperr.ErrInvalidVerificationCode
				return nil
			}
			return err
		}
		valid := pending.Attempts < maxVerificationAttempts && time.Now().Before(pending.ExpiresAt) &&
			hmac.Equal(pending.CodeHash, s.verificationHash(email, code))
		if !valid {
			if pending.Attempts+1 >= maxVerificationAttempts || !time.Now().Before(pending.ExpiresAt) {
				if err := s.store.DeletePendingRegistration(ctx, email); err != nil {
					return err
				}
			} else if err := s.store.IncrementPendingRegistrationAttempts(ctx, email); err != nil {
				return err
			}
			verificationErr = apperr.ErrInvalidVerificationCode
			return nil
		}
		userID, err = s.store.CreateUser(ctx, pending.Email, pending.Username, pending.PasswordHash)
		if err != nil {
			return err
		}
		if err := s.store.CreateProfile(ctx, userID, pending.Username); err != nil {
			return err
		}
		if _, err := s.store.CreateSession(ctx, userID, tokenHash, time.Now().Add(s.sessionTTL)); err != nil {
			return err
		}
		return s.store.DeletePendingRegistration(ctx, email)
	})
	if err != nil {
		return "", nil, err
	}
	if verificationErr != nil {
		return "", nil, verificationErr
	}
	user, err := s.store.GetUserByID(ctx, userID)
	if err != nil {
		return "", nil, err
	}
	return raw, user, nil
}

func (s *Service) verificationHash(email, code string) []byte {
	h := hmac.New(sha256.New, s.verificationSecret)
	h.Write([]byte(email))
	h.Write([]byte{0})
	h.Write([]byte(code))
	return h.Sum(nil)
}

func newVerificationCode() (string, error) {
	var b [4]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	n := (uint32(b[0])<<24 | uint32(b[1])<<16 | uint32(b[2])<<8 | uint32(b[3])) % 1000000
	return fmt.Sprintf("%06d", n), nil
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
