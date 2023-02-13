-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
SET search_path TO devices_api, public;
ALTER TABLE user_device_data ADD COLUMN error_data jsonb;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
SET search_path TO devices_api, public;
ALTER TABLE user_device_data DROP COLUMN error_data;
-- +goose StatementEnd
