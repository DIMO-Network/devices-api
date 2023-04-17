-- +goose Up
-- +goose StatementBegin
SET search_path = devices_api, public;

ALTER TABLE autopi_units
    ADD COLUMN beneficiary bytea;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SET search_path = devices_api, public;

ALTER TABLE autopi_units
    DROP COLUMN beneficiary;
-- +goose StatementEnd
