-- +goose Up
-- +goose StatementBegin

-- These are devices for which we don't yet have a serial.
CREATE TABLE partial_aftermarket_devices(
    token_id numeric(78)
        CONSTRAINT partial_aftermarket_devices_pkey PRIMARY KEY
        CONSTRAINT partial_aftermarket_devices_token_id_check CHECK(token_id > 0),
    manufacturer_token_id numeric(78) NOT NULL
        CONSTRAINT partial_aftermarket_devices_manufacturer_token_id_check CHECK(token_id > 0),
    ethereum_address bytea NOT NULL CONSTRAINT partial_aftermarket_devices_ethereum_address_check CHECK(length(ethereum_address) = 20)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE partial_aftermarket_devices;
-- +goose StatementEnd
