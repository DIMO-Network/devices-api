-- +goose Up
-- +goose StatementBegin
UPDATE devices_api.device_integrations SET country = 'USA' WHERE country = 'us ';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- +goose StatementEnd
