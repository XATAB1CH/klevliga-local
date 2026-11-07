-- name: CreateUser :one
INSERT INTO users (email, username, password_hash)
VALUES ($1, $2, $3)
RETURNING *;

-- name: CreateProfile :one
INSERT INTO user_profiles (user_id, display_name)
VALUES ($1, $2)
RETURNING *;

-- name: RegistrationAvailable :one
SELECT NOT EXISTS (
    SELECT 1 FROM users WHERE email = sqlc.arg('email') OR username = sqlc.arg('username')
);

-- name: UpsertPendingRegistration :exec
INSERT INTO pending_registrations (email, username, password_hash, code_hash, expires_at)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (email) DO UPDATE SET
    username = EXCLUDED.username,
    password_hash = EXCLUDED.password_hash,
    code_hash = EXCLUDED.code_hash,
    attempts = 0,
    expires_at = EXCLUDED.expires_at,
    created_at = now();

-- name: GetPendingRegistrationForUpdate :one
SELECT email, username, password_hash, code_hash, attempts, expires_at
FROM pending_registrations
WHERE email = $1
FOR UPDATE;

-- name: IncrementPendingRegistrationAttempts :exec
UPDATE pending_registrations SET attempts = attempts + 1 WHERE email = $1;

-- name: DeletePendingRegistration :exec
DELETE FROM pending_registrations WHERE email = $1;

-- name: GetUserByEmail :one
SELECT u.id, u.email, u.username, u.password_hash, u.status, u.created_at,
       p.display_name, p.bio, p.region, p.avatar_url, p.role
FROM users u
JOIN user_profiles p ON p.user_id = u.id
WHERE u.email = $1;

-- name: GetUserByID :one
SELECT u.id, u.email, u.username, u.password_hash, u.status, u.created_at,
       p.display_name, p.bio, p.region, p.avatar_url, p.role
FROM users u
JOIN user_profiles p ON p.user_id = u.id
WHERE u.id = $1;

-- name: SearchUsers :many
SELECT u.id, u.username, p.display_name, p.avatar_url, p.region, p.bio, u.created_at
FROM users u
JOIN user_profiles p ON p.user_id = u.id
WHERE u.status = 'active'
  AND (u.username ILIKE '%' || $1 || '%' OR p.display_name ILIKE '%' || $1 || '%')
ORDER BY u.username
LIMIT $2;
