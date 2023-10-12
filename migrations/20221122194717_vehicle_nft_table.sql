-- +goose Up
-- +goose StatementBegin
SET search_path = devices_api, public;

CREATE TABLE vehicle_nfts(
    mint_request_id char(27)
        CONSTRAINT vehicle_nfts_mint_request_id_pkey PRIMARY KEY
        CONSTRAINT vehicle_nfts_mint_request_id_fkey REFERENCES meta_transaction_requests(id),
    user_device_id char(27)
        CONSTRAINT vehicle_nfts_user_device_id_fkey REFERENCES user_devices(id) ON DELETE SET NULL
        CONSTRAINT vehicle_nfts_user_device_id_key UNIQUE,
    vin char(17) NOT NULL, -- Want some constraint on minting the same vin, but it's not clear what.
    token_id numeric(78, 0)
        CONSTRAINT vehicle_nfts_token_id_key UNIQUE,
    owner_address bytea
        CONSTRAINT vehicle_nfts_owner_address_check CHECK (length(owner_address) = 20)
);

INSERT INTO vehicle_nfts (mint_request_id, user_device_id, vin, token_id)
SELECT mint_meta_transaction_request_id, id, vin_identifier, token_id
FROM user_devices
WHERE mint_meta_transaction_request_id IS NOT NULL;

ALTER TABLE user_devices
    DROP COLUMN mint_meta_transaction_request_id,
    DROP COLUMN token_id;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SET search_path = devices_api, public;

ALTER TABLE user_devices
    ADD COLUMN mint_meta_transaction_request_id char(27)
        CONSTRAINT user_devices_mint_meta_transaction_request_id_key UNIQUE
        CONSTRAINT user_devices_mint_meta_transaction_request_id_fkey REFERENCES meta_transaction_requests(id),
    ADD COLUMN token_id numeric(78, 0)
        CONSTRAINT user_devices_token_id_key UNIQUE;

UPDATE user_devices AS ud
    SET
        mint_meta_transaction_request_id = vn.mint_request_id,
        token_id = vn.token_id
    FROM vehicle_nfts AS vn 
    WHERE ud.id = vn.user_device_id;

DROP TABLE vehicle_nfts;
-- +goose StatementEnd
