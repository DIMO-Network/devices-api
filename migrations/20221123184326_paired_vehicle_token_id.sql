-- +goose Up
-- +goose StatementBegin
SET search_path = devices_api, public;

ALTER TABLE autopi_units
    ADD COLUMN vehicle_token_id numeric(78, 0)
        CONSTRAINT autopi_units_vehicle_token_id_key UNIQUE
        CONSTRAINT autopi_units_vehicle_token_id_fkey REFERENCES vehicle_nfts(token_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SET search_path = devices_api, public;

ALTER TABLE autopi_units
    DROP COLUMN vehicle_token_id;
-- +goose StatementEnd
