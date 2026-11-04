-- name: CreateReport :one
INSERT INTO reports (author_id, topic_id, water_id, title, body, fished_at, location_visibility, species)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- name: GetReport :one
SELECT r.*,
       (SELECT p.display_name FROM user_profiles p WHERE p.user_id = r.author_id) AS author_name
FROM reports r
WHERE r.id = $1;

-- name: GetReportByTopic :one
SELECT r.*,
       (SELECT p.display_name FROM user_profiles p WHERE p.user_id = r.author_id) AS author_name
FROM reports r
WHERE r.topic_id = $1;

-- name: ListReports :many
SELECT r.*,
       (SELECT p.display_name FROM user_profiles p WHERE p.user_id = r.author_id) AS author_name
FROM reports r
WHERE (sqlc.narg('water_id')::uuid IS NULL OR r.water_id = sqlc.narg('water_id'))
  AND (
    sqlc.narg('cursor_at')::timestamptz IS NULL
    OR (r.created_at, r.id) < (sqlc.narg('cursor_at')::timestamptz, sqlc.narg('cursor_id')::uuid)
  )
ORDER BY r.created_at DESC, r.id DESC
LIMIT sqlc.arg('limit');

-- name: SearchReports :many
SELECT r.*,
       (SELECT p.display_name FROM user_profiles p WHERE p.user_id = r.author_id) AS author_name
FROM reports r
WHERE r.title ILIKE '%' || $1 || '%' OR r.body ILIKE '%' || $1 || '%'
ORDER BY r.created_at DESC
LIMIT $2;
