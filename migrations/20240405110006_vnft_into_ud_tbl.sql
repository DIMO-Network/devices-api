-- +goose Up
-- +goose StatementBegin
SET search_path = devices_api, public;
CREATE TABLE vehicle_nfts_backup AS SELECT *  FROM vehicle_nfts;

ALTER TABLE user_devices
    ADD COLUMN mint_request_id char(27) CONSTRAINT user_devices_mint_request_id_fkey UNIQUE REFERENCES meta_transaction_requests(id);

ALTER TABLE user_devices
    ADD COLUMN burn_request_id char(27) CONSTRAINT user_devices_burn_request_id_fkey UNIQUE REFERENCES meta_transaction_requests(id);

ALTER TABLE user_devices
    ADD COLUMN token_id numeric(78, 0) CONSTRAINT user_devices_token_id_key UNIQUE;

ALTER TABLE user_devices
    ADD COLUMN claim_id varchar CONSTRAINT user_devices_claim_id_fkey UNIQUE REFERENCES verifiable_credentials(claim_id);

ALTER TABLE user_devices
    ADD COLUMN owner_address BYTEA CONSTRAINT user_devices_owner_address_check CHECK (length(owner_address) = 20);

UPDATE user_devices 
    SET 
        burn_request_id = vnft.burn_request_id,
        mint_request_id = vnft.mint_request_id,
        token_id = vnft.token_id,
        claim_id = vnft.claim_id,
        owner_address = vnft.owner_address
FROM vehicle_nfts vnft 
WHERE vnft.user_device_id = user_devices.id;

ALTER TABLE aftermarket_devices DROP CONSTRAINT autopi_units_vehicle_token_id_fkey;
ALTER TABLE aftermarket_devices ADD CONSTRAINT autopi_units_vehicle_token_id_fkey FOREIGN KEY (vehicle_token_id) REFERENCES user_devices(token_id) ON DELETE CASCADE;

ALTER TABLE synthetic_devices DROP CONSTRAINT fkey_vehicle_token_id;
ALTER TABLE synthetic_devices ADD CONSTRAINT fkey_vehicle_token_id FOREIGN KEY (vehicle_token_id) REFERENCES user_devices(token_id);

DROP TABLE vehicle_nfts;
DROP TABLE vehicle_nfts_backup;

CREATE MATERIALIZED VIEW vehicle_nfts
AS
    SELECT 
            ud.id as user_device_id,
            ud.mint_request_id,
            ud.token_id,
            ud.claim_id,
            ud.owner_address,
            ud.vin_identifier as vin
    FROM user_devices ud
WITH DATA;

create or replace function refresh_vehicle_nfts_mat_view()
returns trigger language plpgsql
as $$
begin
    refresh materialized view devices_api.vehicle_nfts;
    return null;
end $$;

create trigger trigger_vehicle_nfts_refresh
after insert or update or delete or truncate
on devices_api.user_devices for each statement 
execute procedure refresh_vehicle_nfts_mat_view();

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';

SET search_path = devices_api, public;
DROP MATERIALIZED VIEW vehicle_nfts;
CREATE TABLE vehicle_nfts(
    mint_request_id char(27)
        CONSTRAINT vehicle_nfts_mint_request_id_pkey PRIMARY KEY
        CONSTRAINT vehicle_nfts_mint_request_id_fkey REFERENCES meta_transaction_requests(id),
    burn_request_id char(27)
        CONSTRAINT vehicle_nfts_burn_request_id_pkey UNIQUE
        CONSTRAINT vehicle_nfts_burn_request_id_fkey REFERENCES meta_transaction_requests(id),
    user_device_id char(27)
        CONSTRAINT vehicle_nfts_user_device_id_fkey REFERENCES user_devices(id) ON DELETE SET NULL
        CONSTRAINT vehicle_nfts_user_device_id_key UNIQUE,
    vin char(17) NOT NULL,
    token_id numeric(78, 0)
        CONSTRAINT vehicle_nfts_token_id_key UNIQUE,
    owner_address bytea
        CONSTRAINT vehicle_nfts_owner_address_check CHECK (length(owner_address) = 20),
    claim_id varchar
        CONSTRAINT vehicle_nfts_claim_id_key UNIQUE
        CONSTRAINT vehicle_nfts_claim_id_fkey REFERENCES verifiable_credentials(claim_id)
);

INSERT INTO vehicle_nfts 
    (user_device_id, mint_request_id, token_id, claim_id, owner_address, vin)
SELECT 
        ud.id as user_device_id,
        ud.mint_request_id,
        ud.token_id,
        ud.claim_id,
        ud.owner_address,
        ud.vin_identifier as vin
FROM user_devices ud;

ALTER TABLE aftermarket_devices DROP CONSTRAINT autopi_units_vehicle_token_id_fkey;
ALTER TABLE aftermarket_devices ADD CONSTRAINT autopi_units_vehicle_token_id_fkey FOREIGN KEY (vehicle_token_id) REFERENCES vehicle_nfts(token_id) ON DELETE CASCADE;

ALTER TABLE synthetic_devices DROP CONSTRAINT fkey_vehicle_token_id;
ALTER TABLE synthetic_devices ADD CONSTRAINT fkey_vehicle_token_id FOREIGN KEY (vehicle_token_id) REFERENCES vehicle_nfts(token_id);

ALTER TABLE user_devices
    DROP COLUMN mint_request_id;

ALTER TABLE user_devices
    DROP COLUMN token_id;

ALTER TABLE user_devices
    DROP COLUMN claim_id;

ALTER TABLE user_devices
    DROP COLUMN owner_address;

SET search_path = devices_api, public;
ALTER TABLE user_devices
    DROP COLUMN burn_request_id;

-- +goose StatementEnd