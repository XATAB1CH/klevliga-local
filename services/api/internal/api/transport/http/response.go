package http

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/XATAB1CH/klev-liga/internal/api/transport/http/dto"
	"github.com/XATAB1CH/klev-liga/internal/domains/apperr"
)

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)

	_ = json.NewEncoder(w).Encode(v)
}

func writeNoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

// writeError формирует единый формат ошибки {error:{code,message,request_id}}.
func writeError(w http.ResponseWriter, r *http.Request, err error) {
	status, code := mapError(err)

	writeJSON(w, status, dto.ApiError{
		Error: dto.ErrorBody{
			Code:      code,
			Message:   err.Error(),
			RequestID: requestID(r.Context()),
		},
	})
}

func mapError(err error) (status int, code string) {
	switch {
	case errors.Is(err, apperr.ErrNotFound):
		return http.StatusNotFound, "NOT_FOUND"
	case errors.Is(err, apperr.ErrConflict):
		return http.StatusConflict, "CONFLICT"
	case errors.Is(err, apperr.ErrValidation):
		return http.StatusBadRequest, "VALIDATION_ERROR"
	case errors.Is(err, apperr.ErrUnauthorized):
		return http.StatusUnauthorized, "UNAUTHORIZED"
	case errors.Is(err, apperr.ErrForbidden):
		return http.StatusForbidden, "FORBIDDEN"
	case errors.Is(err, apperr.ErrNotImplemented):
		return http.StatusNotImplemented, "NOT_IMPLEMENTED"
	case errors.Is(err, apperr.ErrInvalidCredentials):
		return http.StatusUnauthorized, "INVALID_CREDENTIALS"
	default:
		return http.StatusInternalServerError, "INTERNAL"
	}
}

// --- Cursor helpers ---

func encodeCursor(parts ...string) string {
	return base64.RawURLEncoding.EncodeToString([]byte(strings.Join(parts, "|")))
}

func decodeCursor(s string) ([]string, error) {
	raw, err := base64.RawURLEncoding.DecodeString(s)
	if err != nil {
		return nil, err
	}

	return strings.Split(string(raw), "|"), nil
}

func encodeTimeCursor(t time.Time, id uuid.UUID) string {
	return encodeCursor(t.UTC().Format(time.RFC3339Nano), id.String())
}

func decodeTimeCursor(s string) (time.Time, uuid.UUID, error) {
	parts, err := decodeCursor(s)
	if err != nil || len(parts) != 2 {
		return time.Time{}, uuid.Nil, errors.New("invalid cursor")
	}

	t, err := time.Parse(time.RFC3339Nano, parts[0])
	if err != nil {
		return time.Time{}, uuid.Nil, errors.New("invalid cursor")
	}

	id, err := uuid.Parse(parts[1])
	if err != nil {
		return time.Time{}, uuid.Nil, errors.New("invalid cursor")
	}

	return t, id, nil
}

func encodeWaterCursor(name string, id uuid.UUID) string {
	return encodeCursor(name, id.String())
}

func decodeWaterCursor(s string) (string, uuid.UUID, error) {
	parts, err := decodeCursor(s)
	if err != nil || len(parts) != 2 {
		return "", uuid.Nil, errors.New("invalid cursor")
	}

	id, err := uuid.Parse(parts[1])
	if err != nil {
		return "", uuid.Nil, errors.New("invalid cursor")
	}

	return parts[0], id, nil
}
