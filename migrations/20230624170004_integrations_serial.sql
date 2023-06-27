-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
SET search_path = devices_api, public;

alter table user_device_api_integrations
    rename column autopi_unit_id to serial;

alter table aftermarket_devices drop column autopi_device_id;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
SET search_path = devices_api, public;

alter table user_device_api_integrations
    rename column serial to autopi_unit_id;

alter table aftermarket_devices add column autopi_device_id text;
-- +goose StatementEnd
