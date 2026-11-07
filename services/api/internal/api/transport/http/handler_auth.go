package http

import (
	"errors"
	"io"
	"mime"
	"net/http"

	"github.com/XATAB1CH/klev-liga/internal/api/transport/http/dto"
	"github.com/XATAB1CH/klev-liga/internal/domains/apperr"
	"github.com/XATAB1CH/klev-liga/internal/services/avatar"
)

func (h *Handler) register(w http.ResponseWriter, r *http.Request) {
	var req dto.RegisterRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, r, err)

		return
	}

	req.Email = normalizeText(req.Email)
	req.Username = normalizeText(req.Username)
	if !validEmail(req.Email) || !validTextLength(req.Username, usernameMinLength, usernameMaxLength) || !validTextLength(req.Password, passwordMinLength, passwordMaxLength) {
		writeError(w, r, apperr.ErrValidation)

		return
	}

	if err := h.auth.Register(r.Context(), req.Email, req.Username, req.Password); err != nil {
		writeError(w, r, err)

		return
	}

	w.WriteHeader(http.StatusAccepted)
}

func (h *Handler) verifyRegistration(w http.ResponseWriter, r *http.Request) {
	var req dto.VerifyRegistrationRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, r, err)
		return
	}
	req.Email = normalizeText(req.Email)
	req.Code = normalizeText(req.Code)
	if !validEmail(req.Email) || len(req.Code) != 6 {
		writeError(w, r, apperr.ErrValidation)
		return
	}
	for _, char := range req.Code {
		if char < '0' || char > '9' {
			writeError(w, r, apperr.ErrValidation)
			return
		}
	}
	token, _, err := h.auth.VerifyRegistration(r.Context(), req.Email, req.Code)
	if err != nil {
		writeError(w, r, err)
		return
	}
	h.setSessionCookie(w, token)
	writeNoContent(w)
}

func (h *Handler) login(w http.ResponseWriter, r *http.Request) {
	var req dto.LoginRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, r, err)

		return
	}

	req.Email = normalizeText(req.Email)
	if !validEmail(req.Email) || !validTextLength(req.Password, 1, passwordMaxLength) {
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
		if err := h.auth.Logout(r.Context(), c.Value); err != nil {
			// Preserve the cookie on failure so the client can retry revocation.
			writeError(w, r, err)
			return
		}
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

func (h *Handler) uploadAvatar(w http.ResponseWriter, r *http.Request) {
	user, ok := currentUser(r.Context())
	if !ok {
		writeError(w, r, apperr.ErrUnauthorized)
		return
	}
	contentType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || (contentType != "image/jpeg" && contentType != "image/png") {
		writeError(w, r, apperr.ErrUnsupportedMedia)
		return
	}
	if r.ContentLength > avatar.MaxUploadBytes {
		writeError(w, r, apperr.ErrPayloadTooLarge)
		return
	}
	data, err := io.ReadAll(http.MaxBytesReader(w, r.Body, avatar.MaxUploadBytes))
	if err != nil {
		var sizeErr *http.MaxBytesError
		if errors.As(err, &sizeErr) {
			writeError(w, r, apperr.ErrPayloadTooLarge)
		} else {
			writeError(w, r, err)
		}
		return
	}
	processed, err := avatar.Process(data, contentType)
	if err != nil {
		writeError(w, r, apperr.ErrValidation)
		return
	}
	updated, err := h.auth.UpdateAvatar(r.Context(), user.ID, processed)
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, dto.UserFromDomain(updated))
}

func (h *Handler) getAvatar(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeError(w, r, apperr.ErrValidation)
		return
	}
	data, err := h.auth.GetAvatar(r.Context(), id)
	if err != nil {
		writeError(w, r, err)
		return
	}
	w.Header().Set("Content-Type", "image/jpeg")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}
