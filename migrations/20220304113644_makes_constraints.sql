-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
SET search_path TO devices_api, public;
ALTER TABLE device_definitions DROP CONSTRAINT IF EXISTS device_definitions_make_model_year_key;
ALTER TABLE device_definitions ALTER COLUMN device_make_id SET NOT NULL;
ALTER TABLE device_definitions ALTER COLUMN make DROP NOT NULL;
ALTER TABLE device_definitions ADD CONSTRAINT idx_device_make_id_model_year UNIQUE (device_make_id, model, year);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
SET search_path TO devices_api, public;
ALTER TABLE device_definitions DROP CONSTRAINT idx_device_make_id_model_year;
ALTER TABLE device_definitions ALTER COLUMN device_make_id DROP NOT NULL;
ALTER TABLE device_definitions ALTER COLUMN make SET NOT NULL;
ALTER TABLE device_definitions ADD CONSTRAINT device_definitions_make_model_year_key UNIQUE (make, model, year);
-- +goose StatementEnd
