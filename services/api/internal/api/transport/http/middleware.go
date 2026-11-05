package http

import (
	"log/slog"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/google/uuid"

	"github.com/XATAB1CH/klev-liga/internal/domains/apperr"
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

func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}
