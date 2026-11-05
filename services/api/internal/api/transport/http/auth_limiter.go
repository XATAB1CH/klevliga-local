package http

import (
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/XATAB1CH/klev-liga/internal/api/transport/http/dto"
)

type authAttemptWindow struct {
	started time.Time
	count   int
}

// authLimiter bounds both request volume and concurrent Argon2 work. Its state
// is bounded and lives for one server instance; no background goroutine or
// caller-supplied forwarding header is needed.
type authLimiter struct {
	mu         sync.Mutex
	clients    map[string]authAttemptWindow
	window     time.Duration
	limit      int
	maxClients int
	slots      chan struct{}
	now        func() time.Time
}

func newAuthLimiter() *authLimiter {
	return &authLimiter{
		clients: make(map[string]authAttemptWindow), window: time.Minute,
		limit: 10, maxClients: 1024, slots: make(chan struct{}, 2), now: time.Now,
	}
}

func (l *authLimiter) middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		peer, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			peer = r.RemoteAddr
		}
		if retryAfter := l.takeAttempt(peer); retryAfter > 0 {
			writeAuthLimit(w, r, retryAfter)
			return
		}
		select {
		case l.slots <- struct{}{}:
			defer func() { <-l.slots }()
			next.ServeHTTP(w, r)
		default:
			writeAuthLimit(w, r, time.Second)
		}
	})
}

func (l *authLimiter) takeAttempt(peer string) time.Duration {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := l.now()
	window, exists := l.clients[peer]
	if !exists && len(l.clients) >= l.maxClients {
		for key, entry := range l.clients {
			if !now.Before(entry.started.Add(l.window)) {
				delete(l.clients, key)
			}
		}
		if len(l.clients) >= l.maxClients {
			return l.window
		}
	}
	if !exists || !now.Before(window.started.Add(l.window)) {
		window = authAttemptWindow{started: now}
	}
	if window.count >= l.limit {
		return window.started.Add(l.window).Sub(now)
	}
	window.count++
	l.clients[peer] = window
	return 0
}

func writeAuthLimit(w http.ResponseWriter, r *http.Request, retryAfter time.Duration) {
	seconds := int((retryAfter + time.Second - 1) / time.Second)
	if seconds < 1 {
		seconds = 1
	}
	w.Header().Set("Retry-After", strconv.Itoa(seconds))
	writeJSON(w, http.StatusTooManyRequests, dto.ApiError{Error: dto.ErrorBody{
		Code: "RATE_LIMITED", Message: "too many requests; try again later", RequestID: requestID(r.Context()),
	}})
}
