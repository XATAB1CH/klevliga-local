package http

import (
	"net/http"

	"github.com/XATAB1CH/klev-liga/internal/api/transport/http/dto"
	"github.com/XATAB1CH/klev-liga/internal/domains/apperr"
)

func (h *Handler) flagContent(w http.ResponseWriter, r *http.Request) {
	user, ok := currentUser(r.Context())
	if !ok {
		writeError(w, r, apperr.ErrUnauthorized)

		return
	}

	var req dto.FlagRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, r, err)

		return
	}

	if len(req.Reason) < 10 {
		writeError(w, r, apperr.ErrValidation)

		return
	}

	if _, err := h.moderation.Flag(r.Context(), user.ID, req.TargetType, req.TargetID, req.Reason); err != nil {
		writeError(w, r, err)

		return
	}

	w.WriteHeader(http.StatusCreated)
}
