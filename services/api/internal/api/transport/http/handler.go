package http

import (
	"encoding/json"
	"io"
	"mime"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/XATAB1CH/klev-liga/internal/domains/apperr"
	accessSvc "github.com/XATAB1CH/klev-liga/internal/services/access"
	authSvc "github.com/XATAB1CH/klev-liga/internal/services/auth"
	feedSvc "github.com/XATAB1CH/klev-liga/internal/services/feed"
	forumSvc "github.com/XATAB1CH/klev-liga/internal/services/forum"
	moderationSvc "github.com/XATAB1CH/klev-liga/internal/services/moderation"
	reportSvc "github.com/XATAB1CH/klev-liga/internal/services/report"
	searchSvc "github.com/XATAB1CH/klev-liga/internal/services/search"
	socialSvc "github.com/XATAB1CH/klev-liga/internal/services/social"
	waterSvc "github.com/XATAB1CH/klev-liga/internal/services/water"
)

// Handler объединяет все обработчики REST API и cookie-настройки сессии.
type Handler struct {
	auth       *authSvc.Service
	forum      *forumSvc.Service
	report     *reportSvc.Service
	water      *waterSvc.Service
	feed       *feedSvc.Service
	search     *searchSvc.Service
	social     *socialSvc.Service
	moderation *moderationSvc.Service
	access     *accessSvc.Service

	cookieName   string
	cookieSecure bool
	cookieDomain string
	sessionTTL   time.Duration
}

func NewHandler(
	auth *authSvc.Service,
	forum *forumSvc.Service,
	report *reportSvc.Service,
	water *waterSvc.Service,
	feed *feedSvc.Service,
	search *searchSvc.Service,
	social *socialSvc.Service,
	moderation *moderationSvc.Service,
	access *accessSvc.Service,
	cookieName string,
	cookieSecure bool,
	cookieDomain string,
	sessionTTL time.Duration,
) *Handler {
	return &Handler{
		auth:         auth,
		forum:        forum,
		report:       report,
		water:        water,
		feed:         feed,
		search:       search,
		social:       social,
		moderation:   moderation,
		access:       access,
		cookieName:   cookieName,
		cookieSecure: cookieSecure,
		cookieDomain: cookieDomain,
		sessionTTL:   sessionTTL,
	}
}

// --- helpers ---

func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || mediaType != "application/json" {
		return apperr.ErrValidation
	}
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
	if err := dec.Decode(dst); err != nil {
		return apperr.ErrValidation
	}
	if err := dec.Decode(new(any)); err != io.EOF {
		return apperr.ErrValidation
	}

	return nil
}

func pathID(r *http.Request) (uuid.UUID, error) {
	return uuid.Parse(chi.URLParam(r, "id"))
}

func limitParam(r *http.Request) int32 {
	s := r.URL.Query().Get("limit")
	if s == "" {
		return 20
	}

	n, err := strconv.Atoi(s)
	if err != nil || n < 1 {
		return 20
	}
	if n > 100 {
		return 100
	}

	return int32(n)
}

func timeCursorParam(r *http.Request) (*time.Time, *uuid.UUID, error) {
	s := r.URL.Query().Get("cursor")
	if s == "" {
		return nil, nil, nil
	}

	t, id, err := decodeTimeCursor(s)
	if err != nil {
		return nil, nil, apperr.ErrValidation
	}

	return &t, &id, nil
}

func waterCursorParam(r *http.Request) (*string, *uuid.UUID, error) {
	s := r.URL.Query().Get("cursor")
	if s == "" {
		return nil, nil, nil
	}

	name, id, err := decodeWaterCursor(s)
	if err != nil {
		return nil, nil, apperr.ErrValidation
	}

	return &name, &id, nil
}

func optionalUUIDQuery(r *http.Request, name string) (*uuid.UUID, error) {
	s := r.URL.Query().Get(name)
	if s == "" {
		return nil, nil
	}

	id, err := uuid.Parse(s)
	if err != nil {
		return nil, apperr.ErrValidation
	}

	return &id, nil
}

func optionalStringQuery(r *http.Request, name string) *string {
	s := r.URL.Query().Get(name)
	if s == "" {
		return nil
	}

	return &s
}

func strPtr(s string) *string {
	return &s
}

// --- cookies ---

func (h *Handler) setSessionCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     h.cookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   h.cookieSecure,
		Domain:   h.cookieDomain,
		MaxAge:   int(h.sessionTTL.Seconds()),
		SameSite: http.SameSiteLaxMode,
	})
}

func (h *Handler) clearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     h.cookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   h.cookieSecure,
		Domain:   h.cookieDomain,
		MaxAge:   -1,
		SameSite: http.SameSiteLaxMode,
	})
}
