-- +goose Up
-- +goose StatementBegin
CREATE INDEX follows_target_idx ON follows (target_type, target_id, user_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS follows_target_idx;
-- +goose StatementEnd
