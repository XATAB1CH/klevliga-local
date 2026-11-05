package http

import (
	"net/http"

	"github.com/XATAB1CH/klev-liga/internal/api/transport/http/dto"
	"github.com/XATAB1CH/klev-liga/internal/domains/apperr"
)

func (h *Handler) getFeed(w http.ResponseWriter, r *http.Request) {
	var kind *string

	switch r.URL.Query().Get("kind") {
	case "", "all":
		kind = nil
	case "topic":
		kind = strPtr("topic")
	case "report":
		kind = strPtr("report")
	default:
		writeError(w, r, apperr.ErrValidation)

		return
	}

	cursorAt, cursorID, err := timeCursorParam(r)
	if err != nil {
		writeError(w, r, err)

		return
	}

	limit := limitParam(r)

	items, err := h.feed.Get(r.Context(), kind, cursorAt, cursorID, limit+1)
	if err != nil {
		writeError(w, r, err)

		return
	}

	var next *string
	if len(items) > int(limit) {
		last := items[int(limit)-1].Topic
		next = strPtr(encodeTimeCursor(last.LastActivityAt, last.ID))
		items = items[:limit]
	}

	out := make([]*dto.FeedItem, 0, len(items))
	for _, it := range items {
		fi := &dto.FeedItem{
			Kind:  "topic",
			Topic: dto.TopicFromDomain(it.Topic),
		}

		if it.Report != nil {
			fi.Kind = "report"
			fi.Report = dto.ReportFromDomain(it.Report)
		}

		out = append(out, fi)
	}

	writeJSON(w, http.StatusOK, dto.FeedPage{Items: out, NextCursor: next})
}
