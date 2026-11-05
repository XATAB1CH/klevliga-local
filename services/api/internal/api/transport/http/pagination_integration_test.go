//go:build integration

package http

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/XATAB1CH/klev-liga/internal/storages/postgres/repo"
)

func TestEndpointPaginationHasNoMissingOrRepeatedItems(t *testing.T) {
	a := newIntegrationAPI(t)
	ctx := context.Background()
	var topicID uuid.UUID
	for i := 0; i < 5; i++ {
		r := a.report(t, fmt.Sprintf("Отчёт для пагинации %d", i))
		topicID = r.TopicID
		if _, err := a.provider.Store().CreateNotification(ctx, repo.CreateNotificationParams{
			RecipientID: a.userID, Type: "topic_reply", Title: fmt.Sprintf("Ответ %d", i),
		}); err != nil {
			t.Fatal(err)
		}
	}
	categories, err := a.provider.ForumService().ListCategories(ctx)
	if err != nil || len(categories) == 0 {
		t.Fatalf("categories: %v", err)
	}
	for i := 0; i < 3; i++ {
		if _, err := a.provider.ForumService().CreateTopic(ctx, categories[0].ID, a.userID, fmt.Sprintf("Тема пагинации %d", i), "Содержимое темы"); err != nil {
			t.Fatal(err)
		}
	}
	for i := 0; i < 5; i++ {
		if _, err := a.provider.ForumService().CreatePost(ctx, topicID, a.userID, fmt.Sprintf("Ответ %d", i), nil); err != nil {
			t.Fatal(err)
		}
	}
	// Equal timestamps exercise the UUID tie-breaker as well as the timestamp.
	stamp := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	for _, query := range []string{
		"UPDATE topics SET created_at = $1, last_activity_at = $1",
		"UPDATE topic_posts SET created_at = $1",
		"UPDATE reports SET created_at = $1",
		"UPDATE notifications SET created_at = $1",
	} {
		if _, err := a.pool.Exec(ctx, query, stamp); err != nil {
			t.Fatal(err)
		}
	}
	paths := []string{"/feed", "/feed?kind=topic", "/feed?kind=report", "/forum/topics", "/forum/topics/" + topicID.String() + "/posts", "/reports", "/waters", "/notifications"}
	for _, path := range paths {
		t.Run(path, func(t *testing.T) {
			baseline, cursor := a.pageIDs(t, path, 100, "")
			if len(baseline) < 3 || cursor != nil {
				t.Fatalf("invalid fixture: %d baseline items, cursor %v", len(baseline), cursor)
			}
			for _, limit := range []int{1, 2} {
				var got []uuid.UUID
				next := ""
				for page := 0; ; page++ {
					if page > len(baseline) {
						t.Fatalf("limit %d: cursor never terminated", limit)
					}
					items, cursor := a.pageIDs(t, path, limit, next)
					if len(items) > limit {
						t.Fatalf("limit %d returned %d items", limit, len(items))
					}
					got = append(got, items...)
					if cursor == nil {
						break
					}
					if len(items) == 0 || *cursor == next {
						t.Fatalf("limit %d: cursor made no progress", limit)
					}
					next = *cursor
				}
				if !reflect.DeepEqual(got, baseline) {
					t.Errorf("limit %d: page traversal differs from full result\ngot  %v\nwant %v", limit, got, baseline)
				}
			}
		})
	}
}

func (a *integrationAPI) pageIDs(t *testing.T, path string, limit int, cursor string) ([]uuid.UUID, *string) {
	t.Helper()
	u, err := url.Parse(path)
	if err != nil {
		t.Fatal(err)
	}
	query := u.Query()
	query.Set("limit", fmt.Sprint(limit))
	if cursor != "" {
		query.Set("cursor", cursor)
	}
	u.RawQuery = query.Encode()
	status, data := a.request(t, http.MethodGet, u.String(), nil, a.cookie)
	if status != http.StatusOK {
		t.Fatalf("GET %s: %d %s", u, status, data)
	}
	var page struct {
		Items []struct {
			ID    uuid.UUID `json:"id"`
			Topic *struct {
				ID uuid.UUID `json:"id"`
			} `json:"topic"`
		} `json:"items"`
		NextCursor *string `json:"next_cursor"`
	}
	if err := json.Unmarshal(data, &page); err != nil {
		t.Fatal(err)
	}
	ids := make([]uuid.UUID, 0, len(page.Items))
	for _, item := range page.Items {
		id := item.ID
		if item.Topic != nil {
			id = item.Topic.ID
		}
		if id == uuid.Nil {
			t.Fatalf("GET %s: item missing ID: %s", u, data)
		}
		ids = append(ids, id)
	}
	return ids, page.NextCursor
}
