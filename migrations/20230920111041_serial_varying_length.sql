-- +goose Up
-- +goose StatementBegin
ALTER TABLE aftermarket_devices ALTER COLUMN serial TYPE varchar(36);
UPDATE aftermarket_devices SET serial = trim(trailing from serial);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE aftermarket_devices ALTER COLUMN serial TYPE char(36);
-- +goose StatementEnd
