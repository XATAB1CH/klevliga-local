package http

import (
	"net/http"

	"github.com/XATAB1CH/klev-liga/internal/api/transport/http/dto"
	"github.com/XATAB1CH/klev-liga/internal/domains/apperr"
)

func (h *Handler) listWaters(w http.ResponseWriter, r *http.Request) {
	regionID, err := optionalUUIDQuery(r, "region_id")
	if err != nil {
		writeError(w, r, err)

		return
	}

	q := optionalStringQuery(r, "q")

	cursorName, cursorID, err := waterCursorParam(r)
	if err != nil {
		writeError(w, r, err)

		return
	}

	limit := limitParam(r)

	waters, err := h.water.List(r.Context(), regionID, q, cursorName, cursorID, limit+1)
	if err != nil {
		writeError(w, r, err)

		return
	}

	var next *string
	if len(waters) > int(limit) {
		last := waters[int(limit)-1]
		next = strPtr(encodeWaterCursor(last.Name, last.ID))
		waters = waters[:limit]
	}

	writeJSON(w, http.StatusOK, dto.WaterPage{Items: dto.WatersFromDomain(waters), NextCursor: next})
}

func (h *Handler) getWater(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeError(w, r, apperr.ErrValidation)

		return
	}

	wat, err := h.water.Get(r.Context(), id)
	if err != nil {
		writeError(w, r, err)

		return
	}

	writeJSON(w, http.StatusOK, dto.WaterFromDomain(wat))
}
