-- +goose Up
-- +goose StatementBegin
SET search_path = devices_api, public;

ALTER TABLE verifiable_credentials ADD COLUMN issuance_date timestamptz not null default current_timestamp;
ALTER TABLE verifiable_credentials ADD COLUMN expiration_date timestamptz not null default '2023-12-31 23:59:58';
ALTER TABLE verifiable_credentials ALTER COLUMN expiration_date DROP DEFAULT;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SET search_path = devices_api, public;

ALTER TABLE verifiable_credentials DROP COLUMN issuance_date;
ALTER TABLE verifiable_credentials DROP COLUMN expiration_date;
-- +goose StatementEnd
