package repo

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/XATAB1CH/klev-liga/internal/domains/access"
	"github.com/XATAB1CH/klev-liga/internal/domains/auth"
	"github.com/XATAB1CH/klev-liga/internal/storages/postgres/sqlc"
)

func (s *Store) SearchManagedUsers(ctx context.Context, query string, limit int32) ([]*auth.User, error) {
	rows, err := s.queries(ctx).SearchManagedUsers(ctx, sqlc.SearchManagedUsersParams{Query: query, PageLimit: limit})
	if err != nil {
		return nil, wrapDBError("search managed users", err)
	}
	users := make([]*auth.User, 0, len(rows))
	for _, row := range rows {
		users = append(users, &auth.User{ID: row.ID, Username: row.Username, Status: row.Status,
			CreatedAt: row.CreatedAt, DisplayName: row.DisplayName, Roles: row.Roles})
	}
	return users, nil
}

func (s *Store) LockAccessControl(ctx context.Context) error {
	return wrapDBError("lock access control", s.queries(ctx).LockAccessControl(ctx))
}

func (s *Store) GetUserIDByUsername(ctx context.Context, username string) (uuid.UUID, error) {
	id, err := s.queries(ctx).GetUserIDByUsername(ctx, username)
	return id, wrapDBError("find user by username", err)
}

func (s *Store) CountActiveAdmins(ctx context.Context) (int64, error) {
	n, err := s.queries(ctx).CountActiveAdmins(ctx)
	return n, wrapDBError("count administrators", err)
}

func (s *Store) SetUserRoles(ctx context.Context, userID uuid.UUID, roles []string) error {
	return wrapDBError("set user roles", s.queries(ctx).SetUserRoles(ctx, sqlc.SetUserRolesParams{UserID: userID, Roles: roles}))
}

func (s *Store) SetUserStatus(ctx context.Context, userID uuid.UUID, status string) error {
	return wrapDBError("set user status", s.queries(ctx).SetUserStatus(ctx, sqlc.SetUserStatusParams{ID: userID, Status: status}))
}

func (s *Store) RevokeUserSessions(ctx context.Context, userID uuid.UUID) error {
	return wrapDBError("revoke user sessions", s.queries(ctx).RevokeUserSessions(ctx, userID))
}

func (s *Store) AppendAuditEvent(ctx context.Context, event access.Event) error {
	return wrapDBError("append audit event", s.queries(ctx).InsertAuditEvent(ctx, sqlc.InsertAuditEventParams{
		ActorType: event.ActorType, ActorID: uuidPtrToPg(event.ActorID), Action: event.Action,
		TargetType: event.TargetType, TargetID: event.TargetID, Reason: event.Reason,
		RequestID: event.RequestID, Changes: []byte(event.Changes),
	}))
}

func (s *Store) ListAuditEvents(ctx context.Context, f access.AuditFilter) ([]access.Event, error) {
	optionalTime := func(t *time.Time) pgtype.Timestamptz {
		if t == nil {
			return pgtype.Timestamptz{}
		}
		return timeToPg(*t)
	}
	rows, err := s.queries(ctx).ListAuditEvents(ctx, sqlc.ListAuditEventsParams{
		ActorID: uuidPtrToPg(f.ActorID), TargetID: uuidPtrToPg(f.TargetID), Action: f.Action,
		FromTime: optionalTime(f.From), UntilTime: optionalTime(f.Until),
		BeforeTime: optionalTime(f.BeforeTime), BeforeID: uuidPtrToPg(f.BeforeID), PageLimit: f.Limit,
	})
	if err != nil {
		return nil, wrapDBError("list audit events", err)
	}
	events := make([]access.Event, 0, len(rows))
	for _, row := range rows {
		events = append(events, access.Event{
			ID: row.ID, ActorType: row.ActorType, ActorID: uuidPtr(row.ActorID), Action: row.Action,
			TargetType: row.TargetType, TargetID: row.TargetID, Reason: row.Reason,
			RequestID: row.RequestID, Changes: json.RawMessage(row.Changes), CreatedAt: row.CreatedAt,
		})
	}
	return events, nil
}
