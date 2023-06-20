-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
SET search_path = devices_api, public;

ALTER TABLE autopi_units
RENAME COLUMN autopi_unit_id TO "serial";

ALTER TABLE autopi_units
RENAME COLUMN autopi_device_id TO aftermarket_device_id;

ALTER TABLE autopi_units
RENAME TO aftermarket_units;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
SET search_path = devices_api, public;

ALTER TABLE aftermarket_units
RENAME TO autopi_units;

ALTER TABLE autopi_units
RENAME COLUMN "serial" TO autopi_unit_id;

ALTER TABLE autopi_units
RENAME COLUMN aftermarket_device_id TO autopi_device_id;
-- +goose StatementEnd
