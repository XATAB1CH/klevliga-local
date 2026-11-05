package http

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

func TestAuthLimiterRateWindowAndUntrustedForwardingHeader(t *testing.T) {
	l := newAuthLimiter()
	l.limit = 2
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	l.now = func() time.Time { return now }
	handler := l.middleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusNoContent) }))
	for i, want := range []int{http.StatusNoContent, http.StatusNoContent, http.StatusTooManyRequests} {
		req := httptest.NewRequest(http.MethodPost, "/auth/login", nil)
		req.RemoteAddr = "192.0.2.1:12345"
		req.Header.Set("X-Forwarded-For", "203.0.113."+string(rune('1'+i)))
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != want {
			t.Fatalf("attempt %d: got %d, want %d", i+1, rec.Code, want)
		}
		if want == http.StatusTooManyRequests && rec.Header().Get("Retry-After") != "60" {
			t.Fatalf("missing retry interval: %v", rec.Header())
		}
	}
	now = now.Add(time.Minute)
	if retry := l.takeAttempt("192.0.2.1"); retry != 0 {
		t.Fatalf("expired window rejected request: %v", retry)
	}
}

func TestAuthLimiterBoundsConcurrentRequestsAndReleasesSlots(t *testing.T) {
	l := newAuthLimiter()
	entered := make(chan struct{}, cap(l.slots))
	release := make(chan struct{})
	handler := l.middleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		entered <- struct{}{}
		<-release
		w.WriteHeader(http.StatusNoContent)
	}))
	var requests sync.WaitGroup
	for i := 0; i < cap(l.slots); i++ {
		requests.Add(1)
		go func() {
			defer requests.Done()
			req := httptest.NewRequest(http.MethodPost, "/auth/register", nil)
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)
			if rec.Code != http.StatusNoContent {
				t.Errorf("admitted request returned %d", rec.Code)
			}
		}()
	}
	for i := 0; i < cap(l.slots); i++ {
		select {
		case <-entered:
		case <-time.After(5 * time.Second):
			close(release)
			requests.Wait()
			t.Fatal("admitted requests never reached handler")
		}
	}
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/auth/login", nil))
	if rec.Code != http.StatusTooManyRequests {
		t.Errorf("exhausted concurrency returned %d, want 429", rec.Code)
	}
	close(release)
	requests.Wait()
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/auth/login", nil))
	if rec.Code != http.StatusNoContent {
		t.Errorf("released slots unavailable: %d", rec.Code)
	}
}

func TestAuthLimiterBoundsPeerTableAndReclaimsExpiredEntries(t *testing.T) {
	l := newAuthLimiter()
	l.maxClients = 2
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	l.now = func() time.Time { return now }
	if l.takeAttempt("192.0.2.1") != 0 || l.takeAttempt("192.0.2.2") != 0 {
		t.Fatal("initial peers rejected")
	}
	if l.takeAttempt("192.0.2.3") == 0 || len(l.clients) != l.maxClients {
		t.Fatal("peer table exceeded its bound")
	}
	now = now.Add(time.Minute)
	if l.takeAttempt("192.0.2.3") != 0 || len(l.clients) != 1 {
		t.Fatal("expired peer entries were not reclaimed")
	}
}
