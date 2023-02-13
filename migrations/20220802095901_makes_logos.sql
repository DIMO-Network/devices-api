-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
SET search_path = devices_api, public;

ALTER TABLE device_makes ADD COLUMN logo_url text;
ALTER TABLE device_makes ADD COLUMN oem_platform_name text;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
SET search_path = devices_api, public;

ALTER TABLE device_makes DROP COLUMN logo_url;
ALTER TABLE device_makes DROP COLUMN oem_platform_name;

-- +goose StatementEnd
