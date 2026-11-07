-- +goose Up
-- +goose StatementBegin
CREATE TABLE pending_registrations (
    email         citext PRIMARY KEY,
    username      citext NOT NULL UNIQUE,
    password_hash text NOT NULL,
    code_hash     bytea NOT NULL,
    attempts      integer NOT NULL DEFAULT 0,
    expires_at    timestamptz NOT NULL,
    created_at    timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT pending_registrations_attempts_check CHECK (attempts >= 0)
);
CREATE INDEX pending_registrations_expires_idx ON pending_registrations (expires_at);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS pending_registrations;
-- +goose StatementEnd
