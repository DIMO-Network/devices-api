-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
SET search_path = devices_api, public;
alter table autopi_units rename column autopi_unit_id to serial;
alter table autopi_units add column metadata jsonb;
UPDATE autopi_units
SET metadata = jsonb_set(metadata, '{autopi_device_id}', to_jsonb(autopi_device_id));
alter table autopi_units rename to aftermarket_devices;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
SET search_path = devices_api, public;
alter table aftermarket_devices rename column serial to autopi_unit_id;
alter table aftermarket_devices drop column metadata;
alter table aftermarket_devices rename to autopi_units;
-- +goose StatementEnd

-- todo manually add missing col for migration to work