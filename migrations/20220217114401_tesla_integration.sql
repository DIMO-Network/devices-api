-- +goose Up
-- +goose StatementBegin
SET search_path TO devices_api, public;
ALTER TABLE user_device_api_integrations ALTER COLUMN refresh_expires_at DROP NOT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SET search_path TO devices_api, public;
ALTER TABLE user_device_api_integrations ALTER COLUMN refresh_expires_at SET NOT NULL;
-- +goose StatementEnd
