-- +goose Up
-- +goose StatementBegin
SET search_path = devices_api, public;

ALTER TABLE autopi_units ADD COLUMN owner_address bytea
    CONSTRAINT autopi_units_owner_address_check CHECK (length(owner_address) = 20);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SET search_path = devices_api, public;

ALTER TABLE autopi_units DROP COLUMN owner;
-- +goose StatementEnd
