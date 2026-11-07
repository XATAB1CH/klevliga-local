// Package http — HTTP-транспорт (chi router) REST API.
package http

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
)

type Server struct {
	server *http.Server
}

func NewServer(h *Handler, addr string, log *slog.Logger) *Server {
	r := chi.NewRouter()
	authGuard := newAuthLimiter()

	r.Use(requestIDMiddleware)
	r.Use(recovererMiddleware(log))
	r.Use(loggerMiddleware(log))
	r.Use(authenticateMiddleware(h.auth, h.cookieName))

	r.Route("/api/v1", func(r chi.Router) {
		// Auth — публичные.
		r.With(authGuard.middleware).Post("/auth/register", h.register)
		r.With(authGuard.middleware).Post("/auth/register/verify", h.verifyRegistration)
		r.With(authGuard.middleware).Post("/auth/login", h.login)

		// Auth — требуют сессии.
		r.Group(func(r chi.Router) {
			r.Use(requireAuthMiddleware)
			r.Post("/auth/logout", h.logout)
			r.Get("/auth/me", h.me)
		})

		// Users.
		r.Get("/users/{id}", h.getUser)

		// Feed.
		r.Get("/feed", h.getFeed)

		// Forum.
		r.Get("/forum/categories", h.listCategories)
		r.Get("/forum/topics", h.listTopics)
		r.Post("/forum/topics", h.createTopic)
		r.Get("/forum/topics/{id}", h.getTopic)
		r.Get("/forum/topics/{id}/posts", h.listPosts)
		r.Post("/forum/topics/{id}/posts", h.reply)

		// Reports.
		r.Get("/reports", h.listReports)
		r.Post("/reports", h.createReport)
		r.Get("/reports/{id}", h.getReport)

		// Waters.
		r.Get("/waters", h.listWaters)
		r.Get("/waters/{id}", h.getWater)

		// Search.
		r.Get("/search", h.searchHandler)

		// Follows.
		r.Post("/follows", h.follow)
		r.Delete("/follows/{target_type}/{target_id}", h.unfollow)

		// Notifications.
		r.Get("/notifications", h.listNotifications)

		// Media — заглушка (нет хранилища/таблиц в MVP).
		r.Post("/media/uploads", h.createUpload)
		r.Post("/media/uploads/{id}/complete", h.completeUpload)

		// Moderation.
		r.Post("/moderation/flags", h.flagContent)
	})

	return &Server{
		server: &http.Server{
			Addr:              addr,
			Handler:           r,
			ReadHeaderTimeout: 10 * time.Second,
			ReadTimeout:       15 * time.Second,
			WriteTimeout:      30 * time.Second,
			IdleTimeout:       60 * time.Second,
		},
	}
}

func (s *Server) Start(ctx context.Context) error {
	log := slog.Default()
	log.Info("starting http server", "address", s.server.Addr)

	errCh := make(chan error, 1)

	go func() {
		errCh <- s.server.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		log.Info("context done, stopping http server")

		return s.server.Shutdown(context.Background())
	case err := <-errCh:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}

		return err
	}
}
