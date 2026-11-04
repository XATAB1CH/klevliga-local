-- name: CreateSession :one
INSERT INTO sessions (user_id, token_hash, expires_at)
VALUES ($1, $2, $3)
RETURNING id, user_id, expires_at, created_at;

-- name: GetSessionByTokenHash :one
SELECT id, user_id, expires_at, revoked_at
FROM sessions
WHERE token_hash = $1;

-- name: RevokeSession :exec
UPDATE sessions
SET revoked_at = now()
WHERE token_hash = $1 AND revoked_at IS NULL;
