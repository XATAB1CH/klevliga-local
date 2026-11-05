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

-- name: NotifyTopicSubscribers :exec
WITH recipients AS (
    SELECT author_id AS user_id FROM topics WHERE id = sqlc.arg('topic_id')::uuid
    UNION
    SELECT user_id FROM follows
    WHERE target_type = 'topic' AND target_id = sqlc.arg('topic_id')::uuid
)
INSERT INTO notifications (recipient_id, type, actor_id, target_type, target_id, title, target_url)
SELECT r.user_id,
       CASE WHEN r.user_id = t.author_id THEN 'topic_reply' ELSE 'followed_topic' END,
       sqlc.arg('actor_id')::uuid, 'topic', t.id,
       CASE WHEN r.user_id = t.author_id THEN 'Новый ответ в вашей теме'
            ELSE 'Новый ответ в теме, на которую вы подписаны' END,
       sqlc.arg('target_url')::text
FROM recipients r
JOIN users recipient ON recipient.id = r.user_id AND recipient.status = 'active'
JOIN topics t ON t.id = sqlc.arg('topic_id')::uuid AND t.deleted_at IS NULL
WHERE r.user_id <> sqlc.arg('actor_id')::uuid;

-- name: NotifyWaterSubscribers :exec
INSERT INTO notifications (recipient_id, type, actor_id, target_type, target_id, title, target_url)
SELECT f.user_id, 'followed_water', sqlc.arg('actor_id')::uuid, 'topic', sqlc.arg('topic_id')::uuid,
       'Новый отчёт на водоёме, на который вы подписаны', sqlc.arg('target_url')::text
FROM follows f
JOIN users recipient ON recipient.id = f.user_id AND recipient.status = 'active'
WHERE f.target_type = 'water' AND f.target_id = sqlc.arg('water_id')::uuid
  AND f.user_id <> sqlc.arg('actor_id')::uuid;

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
