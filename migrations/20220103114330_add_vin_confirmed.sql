-- +goose Up
-- +goose StatementBegin
SET search_path = devices_api, public;
ALTER TABLE user_devices ADD COLUMN vin_confirmed BOOLEAN NOT NULL DEFAULT false;
ALTER TYPE user_device_api_integration_status ADD VALUE 'DuplicateIntegration';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SET search_path = devices_api, public;
ALTER TABLE user_devices DROP COLUMN vin_confirmed;
-- Can't remove values from enums.
-- +goose StatementEnd
