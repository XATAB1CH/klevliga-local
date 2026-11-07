package http

import (
	"net/http"
	"time"

	"github.com/XATAB1CH/klev-liga/internal/api/transport/http/dto"
	"github.com/XATAB1CH/klev-liga/internal/domains/access"
	"github.com/XATAB1CH/klev-liga/internal/domains/apperr"
	"github.com/XATAB1CH/klev-liga/internal/domains/auth"
)

func (h *Handler) changeRoles(w http.ResponseWriter, r *http.Request) {
	user, _ := currentUser(r.Context()) // Route requires users.manage.
	id, err := pathID(r)
	if err != nil {
		writeError(w, r, apperr.ErrValidation)
		return
	}
	var req struct {
		Roles  []string `json:"roles"`
		Reason string   `json:"reason"`
	}
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, r, err)
		return
	}
	if err := h.access.ChangeRoles(r.Context(), user.ID, id, req.Roles, req.Reason, requestID(r.Context())); err != nil {
		writeError(w, r, err)
		return
	}
	writeNoContent(w)
}

func (h *Handler) changeUserStatus(w http.ResponseWriter, r *http.Request) {
	user, _ := currentUser(r.Context())
	id, err := pathID(r)
	if err != nil {
		writeError(w, r, apperr.ErrValidation)
		return
	}
	var req struct {
		Status string `json:"status"`
		Reason string `json:"reason"`
	}
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, r, err)
		return
	}
	if err := h.access.ChangeStatus(r.Context(), user.ID, id, req.Status, req.Reason, requestID(r.Context())); err != nil {
		writeError(w, r, err)
		return
	}
	writeNoContent(w)
}

func optionalTimeQuery(r *http.Request, name string) (*time.Time, error) {
	value := r.URL.Query().Get(name)
	if value == "" {
		return nil, nil
	}
	t, err := time.Parse(time.RFC3339Nano, value)
	if err != nil {
		return nil, apperr.ErrValidation
	}
	return &t, nil
}

func (h *Handler) listAuditEvents(w http.ResponseWriter, r *http.Request) {
	user, _ := currentUser(r.Context())
	var f access.AuditFilter
	var err error
	if f.ActorID, err = optionalUUIDQuery(r, "actor_id"); err != nil {
		writeError(w, r, err)
		return
	}
	if f.TargetID, err = optionalUUIDQuery(r, "target_id"); err != nil {
		writeError(w, r, err)
		return
	}
	if f.From, err = optionalTimeQuery(r, "from"); err != nil {
		writeError(w, r, err)
		return
	}
	if f.Until, err = optionalTimeQuery(r, "until"); err != nil {
		writeError(w, r, err)
		return
	}
	if f.BeforeTime, f.BeforeID, err = timeCursorParam(r); err != nil {
		writeError(w, r, err)
		return
	}
	f.Action = r.URL.Query().Get("action")
	limit := limitParam(r)
	f.Limit = limit + 1
	events, err := h.access.ListAudit(r.Context(), user.ID, f)
	if err != nil {
		writeError(w, r, err)
		return
	}
	var next *string
	if len(events) > int(limit) {
		last := events[limit-1]
		next = strPtr(encodeTimeCursor(last.CreatedAt, last.ID))
		events = events[:limit]
	}
	writeJSON(w, http.StatusOK, struct {
		Items      []access.Event `json:"items"`
		NextCursor *string        `json:"next_cursor"`
	}{events, next})
}

func (h *Handler) getAdminUser(w http.ResponseWriter, r *http.Request) {
	user, _ := currentUser(r.Context())
	id, err := pathID(r)
	if err != nil {
		writeError(w, r, apperr.ErrValidation)
		return
	}
	target, err := h.access.GetUser(r.Context(), user.ID, id)
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, managedUserFromDomain(target))
}

type managedUser struct {
	*dto.User
	Roles  []string `json:"roles"`
	Status string   `json:"status"`
}

func managedUserFromDomain(user *auth.User) managedUser {
	return managedUser{dto.UserFromDomain(user), user.Roles, user.Status}
}

func (h *Handler) searchAdminUsers(w http.ResponseWriter, r *http.Request) {
	actor, _ := currentUser(r.Context())
	users, err := h.access.SearchUsers(r.Context(), actor.ID, r.URL.Query().Get("q"), limitParam(r))
	if err != nil {
		writeError(w, r, err)
		return
	}
	items := make([]managedUser, 0, len(users))
	for _, user := range users {
		items = append(items, managedUserFromDomain(user))
	}
	writeJSON(w, http.StatusOK, struct {
		Items []managedUser `json:"items"`
	}{items})
}
