package http

import (
	"net/http"

	"github.com/XATAB1CH/klev-liga/internal/api/transport/http/dto"
	"github.com/XATAB1CH/klev-liga/internal/domains/apperr"
)

func (h *Handler) register(w http.ResponseWriter, r *http.Request) {
	var req dto.RegisterRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, r, err)

		return
	}

	if req.Email == "" || len(req.Username) < 3 || len(req.Username) > 32 || len(req.Password) < 12 {
		writeError(w, r, apperr.ErrValidation)

		return
	}

	if err := h.auth.Register(r.Context(), req.Email, req.Username, req.Password); err != nil {
		writeError(w, r, err)

		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (h *Handler) login(w http.ResponseWriter, r *http.Request) {
	var req dto.LoginRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, r, err)

		return
	}

	if req.Email == "" || req.Password == "" {
		writeError(w, r, apperr.ErrValidation)

		return
	}

	token, _, err := h.auth.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		writeError(w, r, err)

		return
	}

	h.setSessionCookie(w, token)
	writeNoContent(w)
}

func (h *Handler) logout(w http.ResponseWriter, r *http.Request) {
	if c, err := r.Cookie(h.cookieName); err == nil {
		_ = h.auth.Logout(r.Context(), c.Value)
	}

	h.clearSessionCookie(w)
	writeNoContent(w)
}

func (h *Handler) me(w http.ResponseWriter, r *http.Request) {
	user, ok := currentUser(r.Context())
	if !ok {
		writeError(w, r, apperr.ErrUnauthorized)

		return
	}

	writeJSON(w, http.StatusOK, dto.UserFromDomain(user))
}

func (h *Handler) getUser(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeError(w, r, apperr.ErrValidation)

		return
	}

	user, err := h.auth.GetUser(r.Context(), id)
	if err != nil {
		writeError(w, r, err)

		return
	}

	writeJSON(w, http.StatusOK, dto.UserFromDomain(user))
}
