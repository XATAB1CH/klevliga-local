package http

import (
	"net/http"

	"github.com/XATAB1CH/klev-liga/internal/api/transport/http/dto"
	"github.com/XATAB1CH/klev-liga/internal/domains/apperr"
	searchSvc "github.com/XATAB1CH/klev-liga/internal/services/search"
)

func (h *Handler) searchHandler(w http.ResponseWriter, r *http.Request) {
	q := normalizeText(r.URL.Query().Get("q"))
	if !validTextLength(q, 2, queryMaxLength) {
		writeError(w, r, apperr.ErrValidation)

		return
	}

	types := map[string]bool{}
	switch t := r.URL.Query().Get("type"); t {
	case "", searchSvc.TypeAll:
		// поиск по всем типам
	case searchSvc.TypeTopics, searchSvc.TypeReports, searchSvc.TypeWaters, searchSvc.TypeUsers:
		types[t] = true
	default:
		writeError(w, r, apperr.ErrValidation)

		return
	}

	res, err := h.search.Search(r.Context(), q, types, limitParam(r))
	if err != nil {
		writeError(w, r, err)

		return
	}

	writeJSON(w, http.StatusOK, struct {
		Topics  []*dto.Topic  `json:"topics"`
		Reports []*dto.Report `json:"reports"`
		Waters  []*dto.Water  `json:"waters"`
		Users   []*dto.User   `json:"users"`
	}{
		Topics:  dto.TopicsFromDomain(res.Topics),
		Reports: dto.ReportsFromDomain(res.Reports),
		Waters:  dto.WatersFromDomain(res.Waters),
		Users:   dto.UsersFromDomain(res.Users),
	})
}
