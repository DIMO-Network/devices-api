-- +goose Up
-- +goose StatementBegin
SET search_path TO devices_api, public;
ALTER TABLE user_device_data ADD COLUMN last_odometer_event_at timestamptz;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SET search_path TO devices_api, public;
ALTER TABLE user_device_Data DROP COLUMN last_odometer_event_at;
-- +goose StatementEnd
