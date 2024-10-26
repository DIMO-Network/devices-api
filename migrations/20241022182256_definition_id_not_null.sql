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
ALTER TABLE user_devices ALTER COLUMN definition_id DROP NOT NULL;
-- +goose StatementEnd
