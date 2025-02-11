-- +goose Up
-- +goose StatementBegin
SELECT 'down SQL query';
SET search_path = devices_api, public;
drop table dcn;
DROP TABLE user_device_to_geofence;
DROP TABLE geofences;
DROP TYPE geofence_type;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
SET search_path = devices_api, public;
-- +goose StatementEnd
