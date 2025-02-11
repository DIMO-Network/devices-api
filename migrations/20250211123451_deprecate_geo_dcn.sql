-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
SET search_path = devices_api, public;
ALTER TABLE user_devices ALTER COLUMN definition_id SET NOT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
SET search_path = devices_api, public;
drop table dcn;
DROP TABLE user_device_to_geofence;
DROP TABLE geofences;
DROP TYPE geofence_type;
-- +goose StatementEnd
