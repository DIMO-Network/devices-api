-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
ALTER TABLE user_device_data ADD COLUMN real_last_odometer_event_at timestamptz;
-- +goose StatementEnd



-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
ALTER TABLE user_device_data DROP COLUMN real_last_odometer_event_at;
-- +goose StatementEnd
