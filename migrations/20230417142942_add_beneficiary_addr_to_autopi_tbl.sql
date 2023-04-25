-- +goose Up
-- +goose StatementBegin
SET search_path = devices_api, public;

ALTER TABLE autopi_units ADD COLUMN beneficiary bytea;
ALTER TABLE autopi_units ADD CONSTRAINT autopi_units_beneficiary_address_check CHECK (length(beneficiary) = 20) NOT VALID;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SET search_path = devices_api, public;

ALTER TABLE autopi_units DROP COLUMN beneficiary;
-- +goose StatementEnd
