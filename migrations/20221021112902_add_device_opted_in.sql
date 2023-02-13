-- +goose Up
-- +goose StatementBegin
SET search_path = devices_api, public;

ALTER TABLE user_devices ADD COLUMN opted_in_at timestamptz;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SET search_path = devices_api, public;

ALTER TABLE user_devices DROP COLUMN opted_in_at;
-- +goose StatementEnd
