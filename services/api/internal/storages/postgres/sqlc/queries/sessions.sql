-- name: CreateSession :one
INSERT INTO sessions (user_id, token_hash, expires_at)
SELECT u.id, sqlc.arg('token_hash')::bytea, sqlc.arg('expires_at')::timestamptz
FROM users u
WHERE u.id = sqlc.arg('user_id') AND u.status = 'active'
FOR SHARE
RETURNING sessions.id, sessions.user_id, sessions.expires_at, sessions.created_at;

-- name: GetSessionByTokenHash :one
SELECT id, user_id, expires_at, revoked_at
FROM sessions
WHERE token_hash = $1;

-- name: RevokeSession :exec
UPDATE sessions
SET revoked_at = now()
WHERE token_hash = $1 AND revoked_at IS NULL;
