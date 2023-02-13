-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';

SET search_path = devices_api, public;

ALTER TABLE drivly_data RENAME TO external_vin_data;
ALTER TABLE external_vin_data add column pricing_metadata jsonb;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';

SET search_path = devices_api, public;

ALTER TABLE external_vin_data drop column pricing_metadata;
ALTER TABLE external_vin_data RENAME TO drivly_data;

-- +goose StatementEnd
