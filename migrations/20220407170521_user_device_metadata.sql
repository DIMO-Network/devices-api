-- +goose Up
-- +goose StatementBegin
ALTER TABLE user_devices ADD COLUMN metadata jsonb;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE user_devices ADD COLUMN metadata jsonb;
-- +goose StatementEnd
