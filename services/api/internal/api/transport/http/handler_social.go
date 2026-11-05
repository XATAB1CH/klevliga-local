package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/XATAB1CH/klev-liga/internal/api/transport/http/dto"
	"github.com/XATAB1CH/klev-liga/internal/domains/apperr"
)

func (h *Handler) follow(w http.ResponseWriter, r *http.Request) {
	user, ok := currentUser(r.Context())
	if !ok {
		writeError(w, r, apperr.ErrUnauthorized)

		return
	}

	var req dto.FollowRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, r, err)

		return
	}

	if err := h.social.Follow(r.Context(), user.ID, req.TargetType, req.TargetID); err != nil {
		writeError(w, r, err)

		return
	}

	writeNoContent(w)
}

func (h *Handler) unfollow(w http.ResponseWriter, r *http.Request) {
	user, ok := currentUser(r.Context())
	if !ok {
		writeError(w, r, apperr.ErrUnauthorized)

		return
	}

	targetID, err := uuid.Parse(chi.URLParam(r, "target_id"))
	if err != nil {
		writeError(w, r, apperr.ErrValidation)

		return
	}

	if err := h.social.Unfollow(r.Context(), user.ID, chi.URLParam(r, "target_type"), targetID); err != nil {
		writeError(w, r, err)

		return
	}

	writeNoContent(w)
}

func (h *Handler) listNotifications(w http.ResponseWriter, r *http.Request) {
	user, ok := currentUser(r.Context())
	if !ok {
		writeError(w, r, apperr.ErrUnauthorized)

		return
	}

	cursorAt, cursorID, err := timeCursorParam(r)
	if err != nil {
		writeError(w, r, err)

		return
	}

	limit := limitParam(r)

	notifications, err := h.social.ListNotifications(r.Context(), user.ID, cursorAt, cursorID, limit+1)
	if err != nil {
		writeError(w, r, err)

		return
	}

	var next *string
	if len(notifications) > int(limit) {
		last := notifications[int(limit)-1]
		next = strPtr(encodeTimeCursor(last.CreatedAt, last.ID))
		notifications = notifications[:limit]
	}

	writeJSON(w, http.StatusOK, dto.NotificationPage{
		Items:      dto.NotificationsFromDomain(notifications),
		NextCursor: next,
	})
}
