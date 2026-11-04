-- name: Follow :exec
INSERT INTO follows (user_id, target_type, target_id)
VALUES ($1, $2, $3)
ON CONFLICT DO NOTHING;

-- name: Unfollow :exec
DELETE FROM follows
WHERE user_id = $1 AND target_type = $2 AND target_id = $3;

-- name: CreateNotification :one
INSERT INTO notifications (recipient_id, type, actor_id, target_type, target_id, title, target_url)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: ListNotifications :many
SELECT *
FROM notifications
WHERE recipient_id = sqlc.arg('recipient_id')
  AND (
    sqlc.narg('cursor_at')::timestamptz IS NULL
    OR (created_at, id) < (sqlc.narg('cursor_at')::timestamptz, sqlc.narg('cursor_id')::uuid)
  )
ORDER BY created_at DESC, id DESC
LIMIT sqlc.arg('limit');

-- name: CreateAbuseReport :one
INSERT INTO abuse_reports (reporter_id, target_type, target_id, reason)
VALUES ($1, $2, $3, $4)
RETURNING id, status, created_at;
