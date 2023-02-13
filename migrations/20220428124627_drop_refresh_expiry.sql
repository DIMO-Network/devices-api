-- +goose Up
-- +goose StatementBegin
SET search_path TO devices_api, public;
ALTER TABLE user_device_api_integrations DROP COLUMN refresh_expires_at;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SET search_path TO devices_api, public;
ALTER TABLE user_device_api_integrations ADD COLUMN refresh_expires_at timestamptz;
-- +goose StatementEnd
