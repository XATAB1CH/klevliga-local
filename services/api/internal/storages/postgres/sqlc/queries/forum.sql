-- name: ListCategories :many
SELECT id, parent_id, slug, name, position
FROM forum_categories
WHERE status = 'active'
ORDER BY position, name;

-- name: GetCategoryBySlug :one
SELECT id, parent_id, slug, name, position
FROM forum_categories
WHERE slug = $1 AND status = 'active';

-- name: CreateTopic :one
INSERT INTO topics (category_id, author_id, slug, title, body, type)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetTopic :one
SELECT t.*,
       (SELECT p.display_name FROM user_profiles p WHERE p.user_id = t.author_id) AS author_name,
       (SELECT c.name FROM forum_categories c WHERE c.id = t.category_id) AS category_name
FROM topics t
WHERE t.id = $1 AND t.deleted_at IS NULL;

-- name: ListTopics :many
SELECT t.*,
       (SELECT p.display_name FROM user_profiles p WHERE p.user_id = t.author_id) AS author_name,
       (SELECT c.name FROM forum_categories c WHERE c.id = t.category_id) AS category_name
FROM topics t
WHERE t.deleted_at IS NULL
  AND (sqlc.narg('category_id')::uuid IS NULL OR t.category_id = sqlc.narg('category_id'))
  AND (
    sqlc.narg('cursor_at')::timestamptz IS NULL
    OR (t.last_activity_at, t.id) < (sqlc.narg('cursor_at')::timestamptz, sqlc.narg('cursor_id')::uuid)
  )
ORDER BY t.last_activity_at DESC, t.id DESC
LIMIT sqlc.arg('limit');

-- name: ListFeed :many
SELECT t.*,
       (SELECT p.display_name FROM user_profiles p WHERE p.user_id = t.author_id) AS author_name,
       (SELECT c.name FROM forum_categories c WHERE c.id = t.category_id) AS category_name
FROM topics t
WHERE t.deleted_at IS NULL
  AND (sqlc.narg('kind')::text IS NULL OR t.type = sqlc.narg('kind') OR (sqlc.narg('kind') = 'topic' AND t.type <> 'report'))
  AND (
    sqlc.narg('cursor_at')::timestamptz IS NULL
    OR (t.last_activity_at, t.id) < (sqlc.narg('cursor_at')::timestamptz, sqlc.narg('cursor_id')::uuid)
  )
ORDER BY t.last_activity_at DESC, t.id DESC
LIMIT sqlc.arg('limit');

-- name: SearchTopics :many
SELECT t.*,
       (SELECT p.display_name FROM user_profiles p WHERE p.user_id = t.author_id) AS author_name,
       (SELECT c.name FROM forum_categories c WHERE c.id = t.category_id) AS category_name
FROM topics t
WHERE t.deleted_at IS NULL
  AND (t.title ILIKE '%' || sqlc.arg('q') || '%' OR t.body ILIKE '%' || sqlc.arg('q') || '%')
ORDER BY t.last_activity_at DESC
LIMIT sqlc.arg('limit');

-- name: TouchTopicActivity :exec
UPDATE topics
SET last_activity_at = now(),
    replies_count = replies_count + 1,
    updated_at = now()
WHERE id = $1 AND deleted_at IS NULL;

-- name: CreatePost :one
INSERT INTO topic_posts (topic_id, author_id, body, parent_post_id)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: ListPosts :many
SELECT tp.*,
       (SELECT p.display_name FROM user_profiles p WHERE p.user_id = tp.author_id) AS author_name
FROM topic_posts tp
WHERE tp.topic_id = sqlc.arg('topic_id')
  AND tp.deleted_at IS NULL
  AND (
    sqlc.narg('cursor_at')::timestamptz IS NULL
    OR (tp.created_at, tp.id) > (sqlc.narg('cursor_at')::timestamptz, sqlc.narg('cursor_id')::uuid)
  )
ORDER BY tp.created_at, tp.id
LIMIT sqlc.arg('limit');
