-- +goose Up
-- +goose StatementBegin
ALTER TABLE user_device_api_integrations ALTER COLUMN serial TYPE varchar(36);
UPDATE user_device_api_integrations SET serial = trim(trailing from serial);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE user_device_api_integrations ALTER COLUMN serial TYPE char(36);
-- +goose StatementEnd
