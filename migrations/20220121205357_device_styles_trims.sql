-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
SET search_path TO devices_api,public;
DROP INDEX IF EXISTS device_definition_namex;
ALTER TABLE device_styles ADD COLUMN sub_model text NOT NULL;
CREATE UNIQUE INDEX device_definition_name_sub_modelx ON device_styles (device_definition_id, source, name, sub_model);
ALTER TABLE device_definitions DROP COLUMN sub_models;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
SET search_path TO devices_api,public;
DROP INDEX device_definition_name_trimx;
ALTER TABLE device_styles DROP COLUMN trim;
ALTER TABLE device_definitions ADD COLUMN sub_models text[];
-- +goose StatementEnd
