-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
SET search_path = devices_api, public;

alter table user_device_api_integrations
    add column tesla_vehicle_id text;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
SET search_path = devices_api, public;

alter table user_device_api_integrations
    drop column tesla_vehicle_id;

-- +goose StatementEnd
