package http

import (
	"log/slog"
	"net/http"
	"net/url"
	"runtime/debug"
	"time"

	"github.com/google/uuid"

	"github.com/XATAB1CH/klev-liga/internal/domains/apperr"
	authDomain "github.com/XATAB1CH/klev-liga/internal/domains/auth"
	authSvc "github.com/XATAB1CH/klev-liga/internal/services/auth"
)

// requestIDMiddleware присваивает каждому запросу уникальный request_id.
func requestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := uuid.NewString()

		w.Header().Set("X-Request-Id", id)

		next.ServeHTTP(w, r.WithContext(withRequestID(r.Context(), id)))
	})
}

// loggerMiddleware логирует метод, путь, статус и длительность запроса.
func loggerMiddleware(log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}

			next.ServeHTTP(rec, r)

			log.Info("http request",
				"method", r.Method,
				"path", r.URL.Path,
				"status", rec.status,
				"duration", time.Since(start).String(),
				"request_id", requestID(r.Context()),
			)
		})
	}
}

// recovererMiddleware перехватывает panic и возвращает 500.
func recovererMiddleware(log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					log.Error("panic recovered", "panic", rec, "stack", string(debug.Stack()))
					writeError(w, r, apperr.ErrInternal)
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}

// authenticateMiddleware опционально подставляет пользователя из сессионной cookie.
func authenticateMiddleware(auth *authSvc.Service, cookieName string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if c, err := r.Cookie(cookieName); err == nil && c.Value != "" {
				if user, err := auth.Authenticate(r.Context(), c.Value); err == nil {
					r = r.WithContext(withUser(r.Context(), user))
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}

// requireAuthMiddleware требует авторизованного пользователя.
func requireAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := currentUser(r.Context()); !ok {
			writeError(w, r, apperr.ErrUnauthorized)

			return
		}

		next.ServeHTTP(w, r)
	})
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func requirePermission(permission authDomain.Permission) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, ok := currentUser(r.Context())
			if !ok {
				writeError(w, r, apperr.ErrUnauthorized)
				return
			}
			if !user.Can(permission) {
				writeError(w, r, apperr.ErrForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// Browser writes must originate on this host. Requests without Origin remain
// usable by local scripts; JSON content type is required by the handlers.
func adminRequestMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			origin := r.Header.Get("Origin")
			if origin != "" {
				u, err := url.Parse(origin)
				if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host != r.Host || u.User != nil || u.Path != "" || u.RawQuery != "" || u.Fragment != "" {
					writeError(w, r, apperr.ErrForbidden)
					return
				}
			}
			if r.Header.Get("Sec-Fetch-Site") == "cross-site" {
				writeError(w, r, apperr.ErrForbidden)
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}

func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}
