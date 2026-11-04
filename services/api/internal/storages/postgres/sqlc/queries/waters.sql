-- name: ListWaters :many
SELECT w.id, w.slug, w.name, w.kind, w.region_id, w.description, w.created_at
FROM waters w
WHERE w.visibility = 'public'
  AND (sqlc.narg('region_id')::uuid IS NULL OR w.region_id = sqlc.narg('region_id'))
  AND (
    sqlc.narg('q')::text IS NULL
    OR w.name ILIKE '%' || sqlc.narg('q') || '%'
    OR w.slug ILIKE '%' || sqlc.narg('q') || '%'
  )
  AND (
    sqlc.narg('cursor_name')::text IS NULL
    OR (w.name, w.id) > (sqlc.narg('cursor_name')::text, sqlc.narg('cursor_id')::uuid)
  )
ORDER BY w.name, w.id
LIMIT sqlc.arg('limit');

-- name: GetWater :one
SELECT id, slug, name, kind, region_id, description, created_at
FROM waters
WHERE id = $1 AND visibility = 'public';

-- name: GetWaterBySlug :one
SELECT id, slug, name, kind, region_id, description, created_at
FROM waters
WHERE slug = $1 AND visibility = 'public';

-- name: SearchWaters :many
SELECT id, slug, name, kind, region_id, description, created_at
FROM waters
WHERE visibility = 'public'
  AND (name ILIKE '%' || $1 || '%' OR slug ILIKE '%' || $1 || '%')
ORDER BY name
LIMIT $2;
