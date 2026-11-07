-- +goose Up
-- +goose StatementBegin
CREATE TABLE user_avatars (
    user_id uuid PRIMARY KEY REFERENCES users (id) ON DELETE CASCADE,
    image_data bytea NOT NULL,
    updated_at timestamptz NOT NULL DEFAULT now()
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS user_avatars;
-- +goose StatementEnd
