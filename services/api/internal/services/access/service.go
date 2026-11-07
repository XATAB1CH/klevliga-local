// Package access manages privileged changes and their transactional audit trail.
package access

import (
	"context"
	"encoding/json"
	"slices"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"

	accessDomain "github.com/XATAB1CH/klev-liga/internal/domains/access"
	"github.com/XATAB1CH/klev-liga/internal/domains/apperr"
	"github.com/XATAB1CH/klev-liga/internal/domains/auth"
	"github.com/XATAB1CH/klev-liga/internal/storages/postgres/repo"
)

type Service struct{ store *repo.Store }

func NewService(store *repo.Store) *Service { return &Service{store: store} }

func validMetadata(reason, requestID string) bool {
	return utf8.RuneCountInString(reason) >= 3 && utf8.RuneCountInString(reason) <= 1000 &&
		len(requestID) >= 1 && len(requestID) <= 100
}

func (s *Service) require(ctx context.Context, actorID uuid.UUID, permission auth.Permission) error {
	actor, err := s.store.GetUserByID(ctx, actorID)
	if err != nil {
		return err
	}
	if !actor.Can(permission) {
		return apperr.ErrForbidden
	}
	return nil
}

func (s *Service) protectLastAdmin(ctx context.Context, target *auth.User) error {
	if target.Status != auth.UserStatusActive || !target.HasRole(auth.RoleAdmin) {
		return nil
	}
	n, err := s.store.CountActiveAdmins(ctx)
	if err != nil {
		return err
	}
	if n <= 1 {
		return apperr.ErrLastAdmin
	}
	return nil
}

// ChangeRoles checks the actor again after acquiring the shared mutation lock.
// This prevents a concurrent revocation from leaving stale admin privileges.
func (s *Service) ChangeRoles(ctx context.Context, actorID, targetID uuid.UUID, roles []string, reason, requestID string) error {
	roles, ok := auth.NormalizeRoles(roles)
	reason = strings.TrimSpace(reason)
	if !ok || !validMetadata(reason, requestID) {
		return apperr.ErrValidation
	}
	return s.store.WithinTx(ctx, func(ctx context.Context) error {
		if err := s.store.LockAccessControl(ctx); err != nil {
			return err
		}
		if err := s.require(ctx, actorID, auth.ManageUsers); err != nil {
			return err
		}
		target, err := s.store.GetUserByID(ctx, targetID)
		if err != nil {
			return err
		}
		if slices.Equal(target.Roles, roles) {
			return nil
		}
		if !slices.Contains(roles, auth.RoleAdmin) {
			if err := s.protectLastAdmin(ctx, target); err != nil {
				return err
			}
		}
		if err := s.store.SetUserRoles(ctx, targetID, roles); err != nil {
			return err
		}
		return s.record(ctx, &actorID, "user.roles_changed", targetID, reason, requestID, map[string]any{
			"roles": map[string]any{"before": target.Roles, "after": roles},
		})
	})
}

func (s *Service) ChangeStatus(ctx context.Context, actorID, targetID uuid.UUID, status, reason, requestID string) error {
	reason = strings.TrimSpace(reason)
	if (status != auth.UserStatusActive && status != auth.UserStatusBlocked) || !validMetadata(reason, requestID) {
		return apperr.ErrValidation
	}
	return s.store.WithinTx(ctx, func(ctx context.Context) error {
		if err := s.store.LockAccessControl(ctx); err != nil {
			return err
		}
		if err := s.require(ctx, actorID, auth.ManageUsers); err != nil {
			return err
		}
		target, err := s.store.GetUserByID(ctx, targetID)
		if err != nil {
			return err
		}
		if target.Status == status {
			return nil
		}
		if status == auth.UserStatusBlocked {
			if err := s.protectLastAdmin(ctx, target); err != nil {
				return err
			}
		}
		// Update locks the account before revoking sessions. Session creation
		// holds a shared account lock, so an in-flight login either commits
		// before this update and is revoked below, or observes blocked status.
		if err := s.store.SetUserStatus(ctx, targetID, status); err != nil {
			return err
		}
		if status == auth.UserStatusBlocked {
			if err := s.store.RevokeUserSessions(ctx, targetID); err != nil {
				return err
			}
		}
		return s.record(ctx, &actorID, "user.status_changed", targetID, reason, requestID, map[string]any{
			"status": map[string]string{"before": target.Status, "after": status},
		})
	})
}

// BootstrapAdmin has no HTTP route. A local CLI can grant the first active
// administrator to an existing account, only while no active admin exists.
func (s *Service) BootstrapAdmin(ctx context.Context, targetID uuid.UUID, reason, requestID string) error {
	reason = strings.TrimSpace(reason)
	if !validMetadata(reason, requestID) {
		return apperr.ErrValidation
	}
	return s.store.WithinTx(ctx, func(ctx context.Context) error {
		if err := s.store.LockAccessControl(ctx); err != nil {
			return err
		}
		n, err := s.store.CountActiveAdmins(ctx)
		if err != nil {
			return err
		}
		if n != 0 {
			return apperr.ErrConflict
		}
		target, err := s.store.GetUserByID(ctx, targetID)
		if err != nil {
			return err
		}
		if target.Status != auth.UserStatusActive {
			return apperr.ErrForbidden
		}
		roles, _ := auth.NormalizeRoles(append(slices.Clone(target.Roles), auth.RoleAdmin))
		if err := s.store.SetUserRoles(ctx, targetID, roles); err != nil {
			return err
		}
		return s.record(ctx, nil, "admin.bootstrap", targetID, reason, requestID, map[string]any{
			"roles": map[string]any{"before": target.Roles, "after": roles},
		})
	})
}

func (s *Service) record(ctx context.Context, actorID *uuid.UUID, action string, targetID uuid.UUID, reason, requestID string, changes map[string]any) error {
	data, err := json.Marshal(changes)
	if err != nil {
		return err
	}
	actorType := "user"
	if actorID == nil {
		actorType = "bootstrap"
	}
	return s.store.AppendAuditEvent(ctx, accessDomain.Event{
		ActorType: actorType, ActorID: actorID, Action: action, TargetType: "user", TargetID: targetID,
		Reason: reason, RequestID: requestID, Changes: data,
	})
}

func (s *Service) ListAudit(ctx context.Context, actorID uuid.UUID, f accessDomain.AuditFilter) ([]accessDomain.Event, error) {
	if err := s.require(ctx, actorID, auth.ViewAudit); err != nil {
		return nil, err
	}
	if f.Limit < 1 || f.Limit > 101 ||
		(f.From != nil && f.Until != nil && !f.From.Before(*f.Until)) ||
		((f.BeforeTime == nil) != (f.BeforeID == nil)) || len(f.Action) > 100 {
		return nil, apperr.ErrValidation
	}
	return s.store.ListAuditEvents(ctx, f)
}

func (s *Service) GetUser(ctx context.Context, actorID, targetID uuid.UUID) (*auth.User, error) {
	if err := s.require(ctx, actorID, auth.ManageUsers); err != nil {
		return nil, err
	}
	return s.store.GetUserByID(ctx, targetID)
}

func (s *Service) SearchUsers(ctx context.Context, actorID uuid.UUID, query string, limit int32) ([]*auth.User, error) {
	if err := s.require(ctx, actorID, auth.ManageUsers); err != nil {
		return nil, err
	}
	query = strings.TrimSpace(query)
	if utf8.RuneCountInString(query) < 2 || utf8.RuneCountInString(query) > 200 || limit < 1 || limit > 100 {
		return nil, apperr.ErrValidation
	}
	return s.store.SearchManagedUsers(ctx, query, limit)
}
