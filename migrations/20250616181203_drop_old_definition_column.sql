-- +goose Up
-- +goose StatementBegin
SET search_path = devices_api, public;

ALTER TABLE user_devices DROP COLUMN device_definition_id;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SET search_path = devices_api, public;

-- Really no going back on this one.
-- +goose StatementEnd
