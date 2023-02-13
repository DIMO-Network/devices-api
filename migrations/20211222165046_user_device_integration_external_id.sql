-- +goose Up
-- +goose StatementBegin
SET search_path = devices_api, public;

ALTER TABLE user_device_api_integrations ADD COLUMN external_id TEXT;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SET search_path = devices_api, public;

ALTER TABLE user_device_api_integrations DROP COLUMN external_id;
-- +goose StatementEnd
