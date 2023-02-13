-- +goose Up
-- +goose StatementBegin
SET search_path = devices_api, public;
ALTER TYPE user_device_api_integration_status ADD VALUE 'Failed';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SET search_path = devices_api, public;
-- You can't remove values from enums
-- +goose StatementEnd
