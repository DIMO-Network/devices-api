-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
SET search_path = devices_api, public;
alter table external_vin_data add column vincario_metadata jsonb;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
SET search_path = devices_api, public;
alter table external_vin_data drop column vincario_metadata;
-- +goose StatementEnd
