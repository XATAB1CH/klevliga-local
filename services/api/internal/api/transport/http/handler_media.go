package http

import (
	"net/http"

	"github.com/XATAB1CH/klev-liga/internal/domains/apperr"
)

// Media-загрузка в MVP не реализована (нет таблиц media и объектного хранилища).
func (h *Handler) createUpload(w http.ResponseWriter, r *http.Request) {
	writeError(w, r, apperr.ErrNotImplemented)
}

func (h *Handler) completeUpload(w http.ResponseWriter, r *http.Request) {
	writeError(w, r, apperr.ErrNotImplemented)
}
