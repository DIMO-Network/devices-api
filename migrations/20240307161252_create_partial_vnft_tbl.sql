-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';

SET search_path = devices_api, public;

CREATE TABLE partial_vehicle_nfts(
    token_id numeric(78, 0) PRIMARY KEY
        CONSTRAINT partial_vehicle_nfts_token_id_key UNIQUE,
    owner_address bytea
        CONSTRAINT partial_vehicle_nfts_owner_address_check CHECK (length(owner_address) = 20),
    make text,
    model text,
    year numeric,
    synthetic_device bool NOT NULL DEFAULT false,
    integration_token_id numeric(78, 0) UNIQUE,
    synthetic_token_id numeric(78, 0) UNIQUE,
    wallet_child_number int UNIQUE,
    wallet_address bytea UNIQUE
        CONSTRAINT partial_vehicle_nfts_synthetic_wallet_address_check CHECK (length(wallet_address) = 20)
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';

SET search_path = devices_api, public;
DROP TABLE partial_vehicle_nfts;
-- +goose StatementEnd
