-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';

CREATE TABLE synthetic_devices (
    vehicle_token_id numeric(78, 0) NOT NULL,
    integration_token_id numeric(78, 0) NOT NULL,
    mint_request_id char(27) NOT NULL UNIQUE,
    wallet_child_number int NOT NULL UNIQUE,
    wallet_address bytea NOT NULL UNIQUE
        CONSTRAINT synthetic_devices_wallet_address_check CHECK (length(wallet_address) = 20),
    token_id numeric(78, 0) UNIQUE,

    PRIMARY KEY (vehicle_token_id, integration_token_id),
    CONSTRAINT fkey_vehicle_token_id FOREIGN KEY (vehicle_token_id) REFERENCES vehicle_nfts(token_id),
    CONSTRAINT fkey_mint_request_id FOREIGN KEY (mint_request_id) REFERENCES meta_transaction_requests(id)
);

CREATE SEQUENCE synthetic_devices_serial_sequence START 1;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';

DROP TABLE synthetic_devices;

DROP SEQUENCE synthetic_devices_serial_sequence;
-- +goose StatementEnd
