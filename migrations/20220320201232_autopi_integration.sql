-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
SET search_path TO devices_api, public;
ALTER TABLE integrations ADD COLUMN metadata jsonb;
ALTER TABLE integrations ADD CONSTRAINT idx_integrations_type_style_vendor UNIQUE (type, style, vendor);
ALTER TABLE user_device_api_integrations ADD COLUMN metadata jsonb;

ALTER TABLE user_device_api_integrations
    ALTER COLUMN access_token DROP NOT NULL;

ALTER TABLE user_device_api_integrations
    ALTER COLUMN access_expires_at DROP NOT NULL;

ALTER TABLE user_device_api_integrations
    ALTER COLUMN refresh_token DROP NOT NULL;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
SET search_path TO devices_api, public;
ALTER TABLE integrations DROP COLUMN metadata;
ALTER TABLE integrations DROP CONSTRAINT idx_integrations_type_style_vendor;
ALTER TABLE user_device_api_integrations DROP COLUMN metadata;

ALTER TABLE user_device_api_integrations
    ALTER COLUMN access_token SET NOT NULL;

ALTER TABLE user_device_api_integrations
    ALTER COLUMN access_expires_at SET NOT NULL;

ALTER TABLE user_device_api_integrations
    ALTER COLUMN refresh_token SET NOT NULL;

-- +goose StatementEnd
