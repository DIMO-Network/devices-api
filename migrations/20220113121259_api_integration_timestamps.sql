-- +goose Up
-- +goose StatementBegin
SET search_path TO devices_api,public;
ALTER TABLE user_device_api_integrations ADD COLUMN created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP;
ALTER TABLE user_device_api_integrations ADD COLUMN updated_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SET search_path TO devices_api,public;
ALTER TABLE user_device_api_integrations DROP COLUMN created_at;
ALTER TABLE user_device_api_integrations DROP COLUMN updated_at;
-- +goose StatementEnd
