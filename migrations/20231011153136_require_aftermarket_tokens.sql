-- +goose Up
-- +goose StatementBegin
ALTER TABLE aftermarket_devices
    ALTER COLUMN token_id SET NOT NULL,
    ALTER COLUMN device_manufacturer_token_id SET NOT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE aftermarket_devices
    ALTER COLUMN token_id SET NULL,
    ALTER COLUMN device_manufacturer_token_id SET NULL;
-- +goose StatementEnd
