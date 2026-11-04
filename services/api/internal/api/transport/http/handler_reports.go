package http

import (
	"net/http"
	"time"

	"github.com/XATAB1CH/klev-liga/internal/api/transport/http/dto"
	"github.com/XATAB1CH/klev-liga/internal/domains/apperr"
	reportDomain "github.com/XATAB1CH/klev-liga/internal/domains/report"
	reportSvc "github.com/XATAB1CH/klev-liga/internal/services/report"
)

func (h *Handler) listReports(w http.ResponseWriter, r *http.Request) {
	waterID, err := optionalUUIDQuery(r, "water_id")
	if err != nil {
		writeError(w, r, err)

		return
	}

	cursorAt, cursorID, err := timeCursorParam(r)
	if err != nil {
		writeError(w, r, err)

		return
	}

	limit := limitParam(r)

	reports, err := h.report.List(r.Context(), waterID, cursorAt, cursorID, limit+1)
	if err != nil {
		writeError(w, r, err)

		return
	}

	var next *string
	if len(reports) > int(limit) {
		last := reports[int(limit)]
		next = strPtr(encodeTimeCursor(last.CreatedAt, last.ID))
		reports = reports[:limit]
	}

	writeJSON(w, http.StatusOK, dto.ReportPage{Items: dto.ReportsFromDomain(reports), NextCursor: next})
}

func (h *Handler) createReport(w http.ResponseWriter, r *http.Request) {
	user, ok := currentUser(r.Context())
	if !ok {
		writeError(w, r, apperr.ErrUnauthorized)

		return
	}

	var req dto.NewReportRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, r, err)

		return
	}

	if len(req.Title) < 5 || req.Body == "" {
		writeError(w, r, apperr.ErrValidation)

		return
	}

	fishedAt, err := time.Parse("2006-01-02", req.FishedAt)
	if err != nil {
		writeError(w, r, apperr.ErrValidation)

		return
	}

	visibility := req.LocationVisibility
	if visibility == "" {
		visibility = reportDomain.LocationVisibilityWaterOnly
	}
	if visibility != reportDomain.LocationVisibilityWaterOnly && visibility != reportDomain.LocationVisibilityPrivate {
		writeError(w, r, apperr.ErrValidation)

		return
	}

	species := req.Species
	if species == nil {
		species = []string{}
	}

	rep, err := h.report.Create(r.Context(), reportSvc.CreateParams{
		AuthorID:           user.ID,
		WaterID:            req.WaterID,
		Title:              req.Title,
		Body:               req.Body,
		FishedAt:           fishedAt,
		LocationVisibility: visibility,
		Species:            species,
	})
	if err != nil {
		writeError(w, r, err)

		return
	}

	writeJSON(w, http.StatusCreated, dto.ReportFromDomain(rep))
}

func (h *Handler) getReport(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeError(w, r, apperr.ErrValidation)

		return
	}

	rep, err := h.report.Get(r.Context(), id)
	if err != nil {
		writeError(w, r, err)

		return
	}

	writeJSON(w, http.StatusOK, dto.ReportFromDomain(rep))
}
