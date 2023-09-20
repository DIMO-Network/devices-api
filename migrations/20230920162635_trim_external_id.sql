-- +goose Up
-- +goose StatementBegin
UPDATE user_device_api_integrations SET external_id = trim(trailing from external_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- +goose StatementEnd
