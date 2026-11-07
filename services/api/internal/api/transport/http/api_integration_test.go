//go:build integration

package http

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pressly/goose/v3"

	"github.com/XATAB1CH/klev-liga/internal/app/provider"
	"github.com/XATAB1CH/klev-liga/internal/config"
	"github.com/XATAB1CH/klev-liga/internal/domains/apperr"
	reportDomain "github.com/XATAB1CH/klev-liga/internal/domains/report"
	reportSvc "github.com/XATAB1CH/klev-liga/internal/services/report"
	"github.com/XATAB1CH/klev-liga/internal/storages/postgres/repo"
	"github.com/XATAB1CH/klev-liga/migrations"
	"github.com/XATAB1CH/klev-liga/pkg/pg"
)

func TestAvatarUploadRequiresSessionAndPublishesProcessedImage(t *testing.T) {
	a := newIntegrationAPI(t)
	var original bytes.Buffer
	source := image.NewRGBA(image.Rect(0, 0, 24, 12))
	for y := 0; y < 12; y++ {
		for x := 0; x < 24; x++ {
			source.Set(x, y, color.RGBA{R: 110, G: 170, B: 190, A: 255})
		}
	}
	if err := png.Encode(&original, source); err != nil {
		t.Fatal(err)
	}
	upload := func(cookie *http.Cookie, data []byte, contentType string) (*http.Response, []byte) {
		t.Helper()
		req, err := http.NewRequest(http.MethodPut, a.server.URL+"/api/v1/auth/avatar", bytes.NewReader(data))
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Content-Type", contentType)
		if cookie != nil {
			req.AddCookie(cookie)
		}
		resp, err := a.server.Client().Do(req)
		if err != nil {
			t.Fatal(err)
		}
		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			t.Fatal(err)
		}
		return resp, body
	}
	if resp, _ := upload(nil, original.Bytes(), "image/png"); resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("anonymous avatar upload: %d", resp.StatusCode)
	}
	if resp, _ := upload(a.cookie, original.Bytes(), "text/plain"); resp.StatusCode != http.StatusUnsupportedMediaType {
		t.Fatalf("invalid MIME: %d", resp.StatusCode)
	}
	if resp, _ := upload(a.cookie, []byte("invalid image"), "image/png"); resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("invalid bytes: %d", resp.StatusCode)
	}
	if resp, _ := upload(a.cookie, make([]byte, 5<<20+1), "image/png"); resp.StatusCode != http.StatusRequestEntityTooLarge {
		t.Fatalf("oversized image: %d", resp.StatusCode)
	}
	if status, _ := a.request(t, http.MethodGet, "/users/"+a.userID.String()+"/avatar", nil, nil); status != http.StatusNotFound {
		t.Fatalf("missing avatar: %d", status)
	}
	resp, body := upload(a.cookie, original.Bytes(), "image/png")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("upload: %d %s", resp.StatusCode, body)
	}
	var result struct {
		AvatarURL string `json:"avatar_url"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(result.AvatarURL, "/api/v1/users/"+a.userID.String()+"/avatar?v=") {
		t.Fatalf("unexpected avatar URL: %q", result.AvatarURL)
	}
	getResp, err := a.server.Client().Get(a.server.URL + result.AvatarURL)
	if err != nil {
		t.Fatal(err)
	}
	defer getResp.Body.Close()
	if getResp.StatusCode != http.StatusOK || getResp.Header.Get("Content-Type") != "image/jpeg" {
		t.Fatalf("public avatar: status=%d type=%q", getResp.StatusCode, getResp.Header.Get("Content-Type"))
	}
	processed, err := io.ReadAll(getResp.Body)
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := jpeg.Decode(bytes.NewReader(processed))
	if err != nil {
		t.Fatal(err)
	}
	if decoded.Bounds().Dx() != 12 || decoded.Bounds().Dy() != 12 {
		t.Fatalf("cropped image: %v", decoded.Bounds())
	}
	status, profile := a.request(t, http.MethodGet, "/auth/me", nil, a.cookie)
	if status != http.StatusOK || !bytes.Contains(profile, []byte(result.AvatarURL)) {
		t.Fatalf("profile was not updated: %d %s", status, profile)
	}
	otherID, otherCookie := a.newUser(t)
	otherResponse, otherBody := upload(otherCookie, original.Bytes(), "image/png")
	if otherResponse.StatusCode != http.StatusOK || !bytes.Contains(otherBody, []byte(otherID.String())) {
		t.Fatalf("second user upload: %d %s", otherResponse.StatusCode, otherBody)
	}
	status, profile = a.request(t, http.MethodGet, "/auth/me", nil, a.cookie)
	if status != http.StatusOK || !bytes.Contains(profile, []byte(result.AvatarURL)) {
		t.Fatalf("another user's upload changed this profile: %d %s", status, profile)
	}
	if resp, _ := upload(a.cookie, original.Bytes(), "image/png"); resp.StatusCode != http.StatusOK {
		t.Fatalf("replace avatar: %d", resp.StatusCode)
	}
}

type integrationAPI struct {
	pool     *pgxpool.Pool
	provider *provider.Provider
	server   *httptest.Server
	userID   uuid.UUID
	cookie   *http.Cookie
	mail     *captureMailSender
}

type captureMailSender struct {
	mu    sync.Mutex
	codes map[string]string
}

func (s *captureMailSender) SendVerificationCode(_ context.Context, recipient, code string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.codes[recipient] = code
	return nil
}

func (s *captureMailSender) code(recipient string) string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.codes[recipient]
}

// Each test owns a fresh schema; the configured database's existing tables and
// records are never truncated or reused. An explicitly configured but unusable
// database fails the test rather than silently skipping it.
func newIntegrationAPI(t *testing.T) *integrationAPI {
	t.Helper()
	dsn := os.Getenv("KLEVLIGA_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("set KLEVLIGA_TEST_DATABASE_URL to run PostgreSQL integration tests")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	admin, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(admin.Close)
	// Keep database-wide extensions outside the disposable schema so its
	// cleanup cannot remove extension objects used by another schema.
	for _, extension := range []string{"citext", "pgcrypto"} {
		if _, err := admin.Exec(ctx, "CREATE EXTENSION IF NOT EXISTS "+extension+" WITH SCHEMA public"); err != nil {
			t.Fatalf("prepare %s extension: %v", extension, err)
		}
	}
	schema := "klevliga_test_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	quotedSchema := pgx.Identifier{schema}.Sanitize()
	if _, err := admin.Exec(ctx, "CREATE SCHEMA "+quotedSchema); err != nil {
		t.Fatalf("create isolated schema: %v", err)
	}
	t.Cleanup(func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cleanupCancel()
		if _, err := admin.Exec(cleanupCtx, "DROP SCHEMA "+quotedSchema+" CASCADE"); err != nil {
			t.Errorf("drop isolated schema: %v", err)
		}
	})
	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		t.Fatal(err)
	}
	poolConfig.ConnConfig.RuntimeParams["search_path"] = schema + ",public"
	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	if err := pg.MigrateFromPool(ctx, pool, migrations.Migrations, goose.WithTableName(schema+".goose_db_version")); err != nil {
		t.Fatalf("migrate isolated schema: %v", err)
	}
	cfg := &config.Config{Auth: config.Auth{SessionTTL: time.Hour, CookieName: "test_session", VerificationTTL: 10 * time.Minute, VerificationSecret: "integration-test-secret"}}
	mail := &captureMailSender{codes: make(map[string]string)}
	p := provider.NewProvider(cfg, pool, mail)
	h := NewHandler(p.AuthService(), p.ForumService(), p.ReportService(), p.WaterService(), p.FeedService(), p.SearchService(), p.SocialService(), p.ModerationService(), p.AccessService(), cfg.Auth.CookieName, false, "", time.Hour)
	s := NewServer(h, ":0", slog.New(slog.NewTextHandler(io.Discard, nil)), false)
	server := httptest.NewServer(s.server.Handler)
	t.Cleanup(server.Close)
	a := &integrationAPI{pool: pool, provider: p, server: server, mail: mail}
	a.userID, a.cookie = a.newUser(t)
	return a
}

func TestRegistrationRequiresCodeAndCreatesSession(t *testing.T) {
	a := newIntegrationAPI(t)
	email := "verify_" + strings.ReplaceAll(uuid.NewString(), "-", "") + "@example.test"
	username := "angler_" + strings.ReplaceAll(uuid.NewString(), "-", "")[:20]
	status, data := a.request(t, http.MethodPost, "/auth/register", map[string]any{
		"email": email, "username": username, "password": "correct horse battery staple",
		"roles": []string{"user", "admin"}, "role": "admin",
	}, nil)
	if status != http.StatusAccepted {
		t.Fatalf("register: %d %s", status, data)
	}
	if code := a.mail.code(email); len(code) != 6 {
		t.Fatalf("sent code = %q", code)
	}
	if _, err := a.provider.Store().GetUserByEmail(context.Background(), email); !errors.Is(err, apperr.ErrNotFound) {
		t.Fatalf("user exists before verification: %v", err)
	}
	status, _ = a.request(t, http.MethodPost, "/auth/register/verify", map[string]string{"email": email, "code": "999999"}, nil)
	if status != http.StatusBadRequest {
		t.Fatalf("wrong code status = %d", status)
	}

	req, err := http.NewRequest(http.MethodPost, a.server.URL+"/api/v1/auth/register/verify", strings.NewReader(`{"email":"`+email+`","code":"`+a.mail.code(email)+`"}`))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := a.server.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("verify: %d %s", resp.StatusCode, body)
	}
	var sessionCookie *http.Cookie
	for _, cookie := range resp.Cookies() {
		if cookie.Name == "test_session" {
			sessionCookie = cookie
		}
	}
	if sessionCookie == nil || sessionCookie.Value == "" {
		t.Fatal("verification did not set session cookie")
	}
	status, data = a.request(t, http.MethodGet, "/auth/me", nil, sessionCookie)
	if status != http.StatusOK || !strings.Contains(string(data), username) {
		t.Fatalf("me: %d %s", status, data)
	}
	if !bytes.Contains(data, []byte(`"roles":["user"]`)) || !bytes.Contains(data, []byte(`"permissions":[]`)) {
		t.Fatalf("registration granted privileged access: %s", data)
	}
}

func (a *integrationAPI) newUser(t *testing.T) (uuid.UUID, *http.Cookie) {
	t.Helper()
	ctx := context.Background()
	name := "user_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	userID, err := a.provider.Store().CreateUser(ctx, name+"@example.test", name, "unused-test-password")
	if err != nil {
		t.Fatal(err)
	}
	if err := a.provider.Store().CreateProfile(ctx, userID, name); err != nil {
		t.Fatal(err)
	}
	token := uuid.NewString()
	hash := sha256.Sum256([]byte(token))
	if _, err := a.provider.Store().CreateSession(ctx, userID, hash[:], time.Now().Add(time.Hour)); err != nil {
		t.Fatal(err)
	}
	return userID, &http.Cookie{Name: "test_session", Value: token}
}

func (a *integrationAPI) request(t *testing.T, method, path string, payload any, cookie *http.Cookie) (int, []byte) {
	t.Helper()
	var body io.Reader
	if payload != nil {
		encoded, err := json.Marshal(payload)
		if err != nil {
			t.Fatal(err)
		}
		body = strings.NewReader(string(encoded))
	}
	req, err := http.NewRequest(method, a.server.URL+"/api/v1"+path, body)
	if err != nil {
		t.Fatal(err)
	}
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if cookie != nil {
		req.AddCookie(cookie)
	}
	resp, err := a.server.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	return resp.StatusCode, data
}

func (a *integrationAPI) report(t *testing.T, title string) *reportDomain.Report {
	t.Helper()
	r, err := a.provider.ReportService().Create(context.Background(), reportSvc.CreateParams{
		AuthorID: a.userID, Title: title, Body: "Описание рыбалки", FishedAt: time.Now(),
		LocationVisibility: reportDomain.LocationVisibilityWaterOnly, Species: []string{},
	})
	if err != nil {
		t.Fatal(err)
	}
	return r
}

func TestDeletedTopicHidesPostsAndReports(t *testing.T) {
	a := newIntegrationAPI(t)
	ctx := context.Background()
	query := "visibility" + strings.ReplaceAll(uuid.NewString(), "-", "")
	deleted := a.report(t, query+" удалённый отчёт")
	visible := a.report(t, query+" открытый отчёт")
	if _, err := a.provider.ForumService().CreatePost(ctx, deleted.TopicID, a.userID, "Скрываемый ответ", nil); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{"/reports/" + deleted.ID.String(), "/forum/topics/" + deleted.TopicID.String() + "/posts"} {
		if status, data := a.request(t, http.MethodGet, path, nil, nil); status != http.StatusOK {
			t.Fatalf("before deletion GET %s: %d %s", path, status, data)
		}
	}
	if _, err := a.pool.Exec(ctx, "UPDATE topics SET deleted_at = now() WHERE id = $1", deleted.TopicID); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{"/reports/" + deleted.ID.String(), "/forum/topics/" + deleted.TopicID.String(), "/forum/topics/" + deleted.TopicID.String() + "/posts", "/forum/topics/" + uuid.NewString() + "/posts"} {
		if status, data := a.request(t, http.MethodGet, path, nil, nil); status != http.StatusNotFound {
			t.Errorf("GET %s: got %d, want 404: %s", path, status, data)
		}
	}
	for _, path := range []string{"/reports", "/forum/topics", "/feed", "/feed?kind=report", "/search?q=" + url.QueryEscape(query)} {
		status, data := a.request(t, http.MethodGet, path, nil, nil)
		if status != http.StatusOK || !json.Valid(data) {
			t.Fatalf("GET %s: %d %s", path, status, data)
		}
		if strings.Contains(string(data), deleted.ID.String()) || strings.Contains(string(data), deleted.TopicID.String()) {
			t.Errorf("GET %s leaked a deleted report/topic: %s", path, data)
		}
		if !strings.Contains(string(data), visible.ID.String()) && !strings.Contains(string(data), visible.TopicID.String()) {
			t.Errorf("GET %s omitted the visible report/topic: %s", path, data)
		}
	}
	// Repository checks protect callers that bypass the HTTP-level topic check.
	posts, err := a.provider.Store().ListPosts(ctx, repo.ListPostsFilter{TopicID: deleted.TopicID, Limit: 20})
	if err != nil || len(posts) != 0 {
		t.Errorf("repository returned deleted topic posts: %v, %v", posts, err)
	}
	if _, err := a.provider.Store().GetReportByTopic(ctx, deleted.TopicID); !errors.Is(err, apperr.ErrNotFound) {
		t.Errorf("GetReportByTopic: got %v, want not found", err)
	}
}

func TestPostParentAndSubscriberNotifications(t *testing.T) {
	a := newIntegrationAPI(t)
	ctx := context.Background()
	category, err := a.provider.Store().GetCategoryBySlug(ctx, "general")
	if err != nil {
		t.Fatal(err)
	}
	ownerID, _ := a.newUser(t)
	followerID, _ := a.newUser(t)
	actorID, _ := a.newUser(t)
	topic, err := a.provider.ForumService().CreateTopic(ctx, category.ID, ownerID, "Тема для подписчиков", "Начальное сообщение")
	if err != nil {
		t.Fatal(err)
	}
	otherTopic, err := a.provider.ForumService().CreateTopic(ctx, category.ID, ownerID, "Другая тема", "Начальное сообщение")
	if err != nil {
		t.Fatal(err)
	}
	foreignParent, err := a.provider.ForumService().CreatePost(ctx, otherTopic.ID, ownerID, "Ответ в другой теме", nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := a.provider.ForumService().CreatePost(ctx, topic.ID, actorID, "Некорректная ссылка", &foreignParent.ID); !errors.Is(err, apperr.ErrValidation) {
		t.Fatalf("cross-topic parent: got %v, want validation error", err)
	}
	if err := a.provider.SocialService().Follow(ctx, followerID, "topic", topic.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := a.provider.ForumService().CreatePost(ctx, topic.ID, actorID, "Корректный ответ", nil); err != nil {
		t.Fatal(err)
	}

	rows, err := a.pool.Query(ctx, `
		SELECT recipient_id, type, actor_id, target_id, target_url
		FROM notifications
		WHERE target_id = $1
		ORDER BY recipient_id`, topic.ID)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	types := make(map[uuid.UUID]string)
	for rows.Next() {
		var recipientID, notificationActorID, targetID uuid.UUID
		var notificationType, targetURL string
		if err := rows.Scan(&recipientID, &notificationType, &notificationActorID, &targetID, &targetURL); err != nil {
			t.Fatal(err)
		}
		if notificationActorID != actorID || targetID != topic.ID || targetURL != "/forum/topics/"+topic.ID.String() {
			t.Errorf("unexpected notification target: actor=%s target=%s url=%q", notificationActorID, targetID, targetURL)
		}
		types[recipientID] = notificationType
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	if len(types) != 2 || types[ownerID] != "topic_reply" || types[followerID] != "followed_topic" {
		t.Fatalf("notifications: got %#v", types)
	}
	if _, exists := types[actorID]; exists {
		t.Fatal("reply author received a notification about their own reply")
	}
}

func TestModerationRejectsMissingTarget(t *testing.T) {
	a := newIntegrationAPI(t)
	_, err := a.provider.ModerationService().Flag(context.Background(), a.userID, "topic", uuid.New(), "Достаточно подробная причина жалобы")
	if !errors.Is(err, apperr.ErrNotFound) {
		t.Fatalf("got %v, want not found", err)
	}
}
