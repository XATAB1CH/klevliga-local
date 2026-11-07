-- name: LockAccessControl :exec
SELECT id FROM access_control_lock WHERE id = true FOR UPDATE;

-- name: SearchManagedUsers :many
SELECT u.id, u.username, u.status, u.created_at, p.display_name, p.roles
FROM users u JOIN user_profiles p ON p.user_id = u.id
WHERE u.username ILIKE '%' || sqlc.arg('query')::text || '%'
   OR p.display_name ILIKE '%' || sqlc.arg('query')::text || '%'
ORDER BY u.username, u.id
LIMIT sqlc.arg('page_limit');

-- name: CountActiveAdmins :one
SELECT count(*) FROM users u JOIN user_profiles p ON p.user_id = u.id
WHERE u.status = 'active' AND p.roles @> ARRAY['admin']::text[];

-- name: GetUserIDByUsername :one
SELECT id FROM users WHERE username = $1;

-- name: SetUserRoles :exec
UPDATE user_profiles SET roles = $2, updated_at = now() WHERE user_id = $1;

-- name: SetUserStatus :exec
UPDATE users SET status = $2, updated_at = now() WHERE id = $1;

-- name: RevokeUserSessions :exec
UPDATE sessions SET revoked_at = now() WHERE user_id = $1 AND revoked_at IS NULL;

-- name: InsertAuditEvent :exec
INSERT INTO audit_events (actor_type, actor_id, action, target_type, target_id, reason, request_id, changes)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8);

-- name: ListAuditEvents :many
SELECT * FROM audit_events
WHERE (sqlc.narg('actor_id')::uuid IS NULL OR actor_id = sqlc.narg('actor_id'))
  AND (sqlc.narg('target_id')::uuid IS NULL OR target_id = sqlc.narg('target_id'))
  AND (sqlc.arg('action')::text = '' OR action = sqlc.arg('action'))
  AND (sqlc.narg('from_time')::timestamptz IS NULL OR created_at >= sqlc.narg('from_time'))
  AND (sqlc.narg('until_time')::timestamptz IS NULL OR created_at < sqlc.narg('until_time'))
  AND (sqlc.narg('before_time')::timestamptz IS NULL OR
       (created_at, id) < (sqlc.narg('before_time'), sqlc.narg('before_id')::uuid))
ORDER BY created_at DESC, id DESC
LIMIT sqlc.arg('page_limit');
