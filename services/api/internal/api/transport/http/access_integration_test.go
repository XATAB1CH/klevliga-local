//go:build integration

package http

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/XATAB1CH/klev-liga/internal/domains/access"
	"github.com/XATAB1CH/klev-liga/internal/domains/apperr"
	"github.com/XATAB1CH/klev-liga/migrations"
	"github.com/XATAB1CH/klev-liga/pkg/pg"
)

func (a *integrationAPI) bootstrap(t *testing.T) {
	t.Helper()
	if err := a.provider.AccessService().BootstrapAdmin(context.Background(), a.userID, "Initial admin setup", "bootstrap-test"); err != nil {
		t.Fatal(err)
	}
}

func TestRoleAssignmentAuthorizationAuditAndRevocation(t *testing.T) {
	a := newIntegrationAPI(t)
	target, cookie := a.newUser(t)
	path := "/admin/users/" + target.String() + "/roles"
	payload := map[string]any{"roles": []string{"user", "support", "moderator"}, "reason": "Назначение оператора и модератора"}
	for _, c := range []*http.Cookie{nil, a.cookie} {
		want := http.StatusUnauthorized
		if c != nil {
			want = http.StatusForbidden
		}
		if status, data := a.request(t, http.MethodPut, path, payload, c); status != want {
			t.Fatalf("unauthorized assignment: %d %s", status, data)
		}
	}
	a.bootstrap(t)
	if status, data := a.request(t, http.MethodPut, path, payload, a.cookie); status != http.StatusNoContent {
		t.Fatalf("assignment: %d %s", status, data)
	}
	status, data := a.request(t, http.MethodGet, "/auth/me", nil, cookie)
	var session struct {
		Roles, Permissions []string
	}
	if err := json.Unmarshal(data, &session); err != nil || status != http.StatusOK ||
		!reflect.DeepEqual(session.Roles, []string{"user", "moderator", "support"}) ||
		!reflect.DeepEqual(session.Permissions, []string{"content.moderate", "support.handle"}) {
		t.Fatalf("session roles: %d %s (%v)", status, data, err)
	}
	for _, methodPath := range [][2]string{{http.MethodPut, path}, {http.MethodGet, "/admin/audit-events"}, {http.MethodGet, "/admin/users/" + target.String()}} {
		if status, data := a.request(t, methodPath[0], methodPath[1], payload, cookie); status != http.StatusForbidden {
			t.Fatalf("staff gained administrative access: %d %s", status, data)
		}
	}
	// Public profiles must not expose private authorization information.
	_, data = a.request(t, http.MethodGet, "/users/"+target.String(), nil, nil)
	if bytes.Contains(data, []byte("roles")) || bytes.Contains(data, []byte("permissions")) || bytes.Contains(data, []byte("email")) {
		t.Fatalf("public profile leaked private fields: %s", data)
	}
	if status, data := a.request(t, http.MethodPut, path, payload, a.cookie); status != http.StatusNoContent {
		t.Fatalf("idempotent assignment: %d %s", status, data)
	}
	var count int
	if err := a.pool.QueryRow(context.Background(), "SELECT count(*) FROM audit_events WHERE action='user.roles_changed'").Scan(&count); err != nil || count != 1 {
		t.Fatalf("no-op duplicated audit: %d (%v)", count, err)
	}
	var changes []byte
	var requestID, reason string
	var actor, savedTarget uuid.UUID
	if err := a.pool.QueryRow(context.Background(), "SELECT actor_id, target_id, request_id, reason, changes FROM audit_events WHERE action='user.roles_changed'").Scan(&actor, &savedTarget, &requestID, &reason, &changes); err != nil {
		t.Fatal(err)
	}
	if actor != a.userID || savedTarget != target || requestID == "" || reason != payload["reason"] || !bytes.Contains(changes, []byte("before")) || bytes.Contains(changes, []byte("password")) {
		t.Fatalf("invalid audit: %s %s %q %q %s", actor, savedTarget, requestID, reason, changes)
	}
	// Grant, then revoke using the very same session: permissions are not cached.
	for _, roles := range [][]string{{"user", "admin"}, {"user"}} {
		payload["roles"] = roles
		if status, data := a.request(t, http.MethodPut, path, payload, a.cookie); status != http.StatusNoContent {
			t.Fatalf("change roles: %d %s", status, data)
		}
		want := http.StatusOK
		if len(roles) == 1 {
			want = http.StatusForbidden
		}
		if status, data := a.request(t, http.MethodGet, "/admin/audit-events", nil, cookie); status != want {
			t.Fatalf("revocation/grant with existing session: %d %s", status, data)
		}
	}
}

func TestLastActiveAdminAndBootstrapAreProtected(t *testing.T) {
	a := newIntegrationAPI(t)
	a.bootstrap(t)
	other, _ := a.newUser(t)
	if err := a.provider.AccessService().BootstrapAdmin(context.Background(), other, "Second bootstrap must fail", "test"); !errors.Is(err, apperr.ErrConflict) {
		t.Fatalf("second bootstrap: %v", err)
	}
	for _, tc := range []struct {
		path string
		body map[string]any
	}{
		{"roles", map[string]any{"roles": []string{"user"}, "reason": "Remove sole administrator"}},
		{"status", map[string]any{"status": "blocked", "reason": "Block sole administrator"}},
	} {
		status, data := a.request(t, http.MethodPut, "/admin/users/"+a.userID.String()+"/"+tc.path, tc.body, a.cookie)
		if status != http.StatusConflict || !bytes.Contains(data, []byte("LAST_ADMIN_REQUIRED")) {
			t.Fatalf("last admin protection: %d %s", status, data)
		}
	}
	var count int
	if err := a.pool.QueryRow(context.Background(), "SELECT count(*) FROM audit_events").Scan(&count); err != nil || count != 1 {
		t.Fatalf("failed mutations wrote audit: %d %v", count, err)
	}
}

func TestConcurrentDemotionsPreserveOneActiveAdmin(t *testing.T) {
	a := newIntegrationAPI(t)
	a.bootstrap(t)
	other, _ := a.newUser(t)
	svc := a.provider.AccessService()
	ctx := context.Background()
	if err := svc.ChangeRoles(ctx, a.userID, other, []string{"user", "admin"}, "Second administrator", "grant"); err != nil {
		t.Fatal(err)
	}
	start := make(chan struct{})
	errs := make(chan error, 2)
	var wg sync.WaitGroup
	for _, id := range []uuid.UUID{a.userID, other} {
		wg.Go(func() {
			<-start
			errs <- svc.ChangeRoles(ctx, id, id, []string{"user"}, "Concurrent self demotion", uuid.NewString())
		})
	}
	close(start)
	wg.Wait()
	close(errs)
	var successes, refused int
	for err := range errs {
		if err == nil {
			successes++
		} else if errors.Is(err, apperr.ErrLastAdmin) {
			refused++
		} else {
			t.Fatal(err)
		}
	}
	n, err := a.provider.Store().CountActiveAdmins(ctx)
	if err != nil || n != 1 || successes != 1 || refused != 1 {
		t.Fatalf("admins=%d success=%d refused=%d err=%v", n, successes, refused, err)
	}
}

func TestConcurrentBootstrapAssignsOnlyOneAdministrator(t *testing.T) {
	a := newIntegrationAPI(t)
	other, _ := a.newUser(t)
	svc := a.provider.AccessService()
	start := make(chan struct{})
	errs := make(chan error, 2)
	for _, id := range []uuid.UUID{a.userID, other} {
		go func() {
			<-start
			errs <- svc.BootstrapAdmin(context.Background(), id, "Concurrent initial setup", uuid.NewString())
		}()
	}
	close(start)
	var successes, conflicts int
	for i := 0; i < 2; i++ {
		err := <-errs
		if err == nil {
			successes++
		} else if errors.Is(err, apperr.ErrConflict) {
			conflicts++
		} else {
			t.Fatal(err)
		}
	}
	if successes != 1 || conflicts != 1 {
		t.Fatalf("successes=%d conflicts=%d", successes, conflicts)
	}
}

func TestAdministrativeSearchIncludesBlockedUsersAndRejectsForeignOrigins(t *testing.T) {
	a := newIntegrationAPI(t)
	a.bootstrap(t)
	target, _ := a.newUser(t)
	ctx := context.Background()
	svc := a.provider.AccessService()
	if err := svc.ChangeStatus(ctx, a.userID, target, "blocked", "Search blocked account", "search-test"); err != nil {
		t.Fatal(err)
	}
	user, err := a.provider.Store().GetUserByID(ctx, target)
	if err != nil {
		t.Fatal(err)
	}
	status, data := a.request(t, http.MethodGet, "/admin/users?q="+url.QueryEscape(user.Username), nil, a.cookie)
	if status != http.StatusOK || !bytes.Contains(data, []byte(target.String())) || !bytes.Contains(data, []byte(`"status":"blocked"`)) || bytes.Contains(data, []byte("password_hash")) {
		t.Fatalf("admin search: %d %s", status, data)
	}
	if status, data := a.request(t, http.MethodGet, "/admin/users?q=x", nil, a.cookie); status != http.StatusBadRequest {
		t.Fatalf("short admin search accepted: %d %s", status, data)
	}
	for _, tc := range []struct {
		origin, site string
		want         int
	}{
		{"https://attacker.invalid", "", http.StatusForbidden},
		{"null", "", http.StatusForbidden},
		{"", "cross-site", http.StatusForbidden},
		{a.server.URL, "same-origin", http.StatusNoContent},
	} {
		req, err := http.NewRequest(http.MethodPut, a.server.URL+"/api/v1/admin/users/"+target.String()+"/roles", strings.NewReader(`{"roles":["user","support"],"reason":"Origin validation test"}`))
		if err != nil {
			t.Fatal(err)
		}
		req.AddCookie(a.cookie)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Origin", tc.origin)
		req.Header.Set("Sec-Fetch-Site", tc.site)
		resp, err := a.server.Client().Do(req)
		if err != nil {
			t.Fatal(err)
		}
		resp.Body.Close()
		if resp.StatusCode != tc.want || resp.Header.Get("Cache-Control") != "no-store" {
			t.Fatalf("origin %q site %q: got %d want %d", tc.origin, tc.site, resp.StatusCode, tc.want)
		}
	}
}

func TestAuditFailureRollsBackRoleStatusAndSessions(t *testing.T) {
	a := newIntegrationAPI(t)
	a.bootstrap(t)
	target, cookie := a.newUser(t)
	ctx := context.Background()
	// Fail the audit INSERT after the actual mutation has happened in the tx.
	_, err := a.pool.Exec(ctx, `CREATE FUNCTION fail_audit_insert() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN RAISE EXCEPTION 'simulated audit failure'; END; $$;
CREATE TRIGGER fail_audit_insert BEFORE INSERT ON audit_events FOR EACH ROW EXECUTE FUNCTION fail_audit_insert();`)
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		path string
		body map[string]any
	}{
		{"roles", map[string]any{"roles": []string{"user", "admin"}, "reason": "Must roll back roles"}},
		{"status", map[string]any{"status": "blocked", "reason": "Must roll back status and revocation"}},
	} {
		if status, data := a.request(t, http.MethodPut, "/admin/users/"+target.String()+"/"+tc.path, tc.body, a.cookie); status != http.StatusInternalServerError {
			t.Fatalf("audit failure: %d %s", status, data)
		}
	}
	user, err := a.provider.Store().GetUserByID(ctx, target)
	if err != nil || user.Status != "active" || !reflect.DeepEqual(user.Roles, []string{"user"}) {
		t.Fatalf("mutation escaped rollback: %#v %v", user, err)
	}
	if status, data := a.request(t, http.MethodGet, "/auth/me", nil, cookie); status != http.StatusOK {
		t.Fatalf("session escaped rollback: %d %s", status, data)
	}
}

func TestBlockingRevokesEverySessionAndUnblockDoesNotRestoreThem(t *testing.T) {
	a := newIntegrationAPI(t)
	a.bootstrap(t)
	target, cookie := a.newUser(t)
	secondToken := uuid.NewString()
	secondHash := sha256.Sum256([]byte(secondToken))
	if _, err := a.provider.Store().CreateSession(context.Background(), target, secondHash[:], time.Now().Add(time.Hour)); err != nil {
		t.Fatal(err)
	}
	for _, state := range []string{"blocked", "active"} {
		if status, data := a.request(t, http.MethodPut, "/admin/users/"+target.String()+"/status", map[string]string{"status": state, "reason": "Account status test"}, a.cookie); status != http.StatusNoContent {
			t.Fatalf("status change: %d %s", status, data)
		}
		for _, sessionCookie := range []*http.Cookie{cookie, {Name: "test_session", Value: secondToken}} {
			if status, data := a.request(t, http.MethodGet, "/auth/me", nil, sessionCookie); status != http.StatusUnauthorized {
				t.Fatalf("revoked session worked: %d %s", status, data)
			}
		}
	}
	var unrevoked int
	if err := a.pool.QueryRow(context.Background(), "SELECT count(*) FROM sessions WHERE user_id=$1 AND revoked_at IS NULL", target).Scan(&unrevoked); err != nil || unrevoked != 0 {
		t.Fatalf("sessions left active: %d %v", unrevoked, err)
	}
}

func TestInFlightLoginCannotLeaveSessionAliveAfterBlock(t *testing.T) {
	a := newIntegrationAPI(t)
	a.bootstrap(t)
	target, _ := a.newUser(t)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	loginTx, err := a.pool.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer loginTx.Rollback(context.Background())
	// Session insertion holds the account lock until the login commits.
	token := sha256.Sum256([]byte(uuid.NewString()))
	if _, err := a.provider.Store().CreateSession(pg.WithTx(ctx, loginTx), target, token[:], time.Now().Add(time.Hour)); err != nil {
		t.Fatal(err)
	}
	var transactionID string
	if err := loginTx.QueryRow(ctx, "SELECT pg_current_xact_id()::text").Scan(&transactionID); err != nil {
		t.Fatal(err)
	}
	blocked := make(chan error, 1)
	go func() {
		blocked <- a.provider.AccessService().ChangeStatus(ctx, a.userID, target, "blocked", "Block concurrent login", "concurrent-block")
	}()
	// Ensure the block really overlaps the uncommitted session. Avoid a timing
	// assumption: wait until the blocking transaction requests our account lock.
	for {
		var waiting bool
		if err := a.pool.QueryRow(ctx, "SELECT EXISTS (SELECT 1 FROM pg_locks WHERE NOT granted AND locktype='transactionid' AND transactionid::text=$1)", transactionID).Scan(&waiting); err != nil {
			t.Fatal(err)
		}
		if waiting {
			break
		}
		select {
		case err := <-blocked:
			t.Fatalf("block completed before the in-flight login committed: %v", err)
		case <-ctx.Done():
			t.Fatal("block did not wait for in-flight login")
		case <-time.After(10 * time.Millisecond):
		}
	}
	if err := loginTx.Commit(ctx); err != nil {
		t.Fatal(err)
	}
	if err := <-blocked; err != nil {
		t.Fatal(err)
	}
	var activeSessions int
	if err := a.pool.QueryRow(ctx, "SELECT count(*) FROM sessions WHERE user_id=$1 AND revoked_at IS NULL", target).Scan(&activeSessions); err != nil || activeSessions != 0 {
		t.Fatalf("in-flight login survived the block: %d %v", activeSessions, err)
	}
	if _, err := a.provider.Store().CreateSession(ctx, target, []byte(uuid.NewString()), time.Now().Add(time.Hour)); !errors.Is(err, apperr.ErrForbidden) {
		t.Fatalf("session could be created while blocked: %v", err)
	}
}

func TestAuditPaginationFiltersAndAppendOnly(t *testing.T) {
	a := newIntegrationAPI(t)
	a.bootstrap(t)
	target, _ := a.newUser(t)
	ctx := context.Background()
	for i := 0; i < 6; i++ {
		roles := []string{"user"}
		if i%2 == 0 {
			roles = append(roles, "support")
		}
		if err := a.provider.AccessService().ChangeRoles(ctx, a.userID, target, roles, "Audit pagination", fmt.Sprint(i)); err != nil {
			t.Fatal(err)
		}
	}
	path := "/admin/audit-events?target_id=" + target.String() + "&actor_id=" + a.userID.String() + "&action=user.roles_changed&limit=2"
	seen := map[uuid.UUID]bool{}
	cursor := ""
	for i := 0; i < 5; i++ {
		status, data := a.request(t, http.MethodGet, path+"&cursor="+url.QueryEscape(cursor), nil, a.cookie)
		var page struct {
			Items      []access.Event `json:"items"`
			NextCursor *string        `json:"next_cursor"`
		}
		if err := json.Unmarshal(data, &page); err != nil || status != http.StatusOK {
			t.Fatalf("audit page: %d %s %v", status, data, err)
		}
		for _, event := range page.Items {
			if seen[event.ID] || event.TargetID != target || event.ActorID == nil || *event.ActorID != a.userID {
				t.Fatalf("invalid or duplicate event: %#v", event)
			}
			seen[event.ID] = true
		}
		if page.NextCursor == nil {
			break
		}
		cursor = *page.NextCursor
	}
	if len(seen) != 6 {
		t.Fatalf("lost audit events: %d", len(seen))
	}
	for _, query := range []string{"UPDATE audit_events SET reason='tamper'", "DELETE FROM audit_events", "TRUNCATE audit_events"} {
		if _, err := a.pool.Exec(ctx, query); err == nil {
			t.Fatalf("allowed audit mutation: %s", query)
		}
	}
	for _, suffix := range []string{"?cursor=bad", "?actor_id=no", "?target_id=no", "?from=bad", "?until=bad", "?from=2026-11-07T12:00:00Z&until=2026-11-06T12:00:00Z"} {
		if status, data := a.request(t, http.MethodGet, "/admin/audit-events"+suffix, nil, a.cookie); status != http.StatusBadRequest {
			t.Fatalf("invalid filter accepted: %s %d %s", suffix, status, data)
		}
	}
	future := time.Now().Add(time.Hour).UTC().Format(time.RFC3339Nano)
	status, data := a.request(t, http.MethodGet, "/admin/audit-events?from="+url.QueryEscape(future), nil, a.cookie)
	if status != http.StatusOK || !bytes.Contains(data, []byte(`"items":[]`)) {
		t.Fatalf("time filter ignored: %d %s", status, data)
	}
}

func TestLegacyRolesMigrationPreservesPrivileges(t *testing.T) {
	a := newIntegrationAPI(t)
	other, _ := a.newUser(t)
	ctx := context.Background()
	source, err := migrations.Migrations.ReadFile("20261107200000_add_roles_and_audit.sql")
	if err != nil {
		t.Fatal(err)
	}
	parts := strings.Split(string(source), "-- +goose Down")
	if _, err := a.pool.Exec(ctx, parts[1]); err != nil {
		t.Fatal(err)
	}
	if _, err := a.pool.Exec(ctx, "UPDATE user_profiles SET role='moderator' WHERE user_id=$1", a.userID); err != nil {
		t.Fatal(err)
	}
	if _, err := a.pool.Exec(ctx, "UPDATE user_profiles SET role='admin' WHERE user_id=$1", other); err != nil {
		t.Fatal(err)
	}
	if _, err := a.pool.Exec(ctx, parts[0]); err != nil {
		t.Fatal(err)
	}
	for id, want := range map[uuid.UUID][]string{a.userID: {"user", "moderator"}, other: {"user", "admin"}} {
		user, err := a.provider.Store().GetUserByID(ctx, id)
		if err != nil || !reflect.DeepEqual(user.Roles, want) {
			t.Fatalf("migration lost privileges: %#v %v", user, err)
		}
	}
	for _, roles := range [][]string{{"admin"}, {"user", "user"}, {"user", "owner"}, {"user", ""}} {
		if _, err := a.pool.Exec(ctx, "UPDATE user_profiles SET roles=$1 WHERE user_id=$2", roles, a.userID); err == nil {
			t.Fatalf("invalid roles accepted by database: %v", roles)
		}
	}
}
