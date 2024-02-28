-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';

SET search_path = devices_api, public;
-- CREATE TABLE user_devices_backup AS SELECT * FROM user_devices;
-- CREATE TABLE vehicle_nfts_backup AS SELECT * FROM vehicle_nfts;


--  create new tbl with combined columns
CREATE TABLE user_devices_v2 (
    -- user device id 
    id char(27) PRIMARY KEY, -- source of truth, a minted vehicle cannot exist without this; this can exist without a minted vehicle

    token_id numeric(78, 0)
        CONSTRAINT vehicles_token_id_key UNIQUE,
    owner_address bytea
        CONSTRAINT vehicles_owner_address_check CHECK (length(owner_address) = 20),
    mint_request_id char(27)
        CONSTRAINT vehicles_mint_request_id_fkey REFERENCES meta_transaction_requests(id),
    
    user_id text, -- this becomes nullable
    vin char(17), -- NOT NULL (?) want constraint on minting the same vin but don't think we have the ability to set it here? 
    vin_confirmed BOOLEAN NOT NULL DEFAULT false,
    claim_id varchar
        CONSTRAINT user_devices_v2_claim_id_fkey UNIQUE REFERENCES verifiable_credentials(claim_id),

    metadata           jsonb,
    name text, -- name the user can give
    custom_image_url text,
    country_code char(3),
    device_style_id char(27),
    opted_in_at timestamptz,

    created_at           timestamptz not null default current_timestamp,
    updated_at           timestamptz not null default current_timestamp
);

-- copy data from existing tbls into new one
INSERT INTO user_devices_v2 
    (
        id, 
        token_id, 
        owner_address, 
        mint_request_id, 
        user_id, 
        vin, 
        vin_confirmed, 
        claim_id,
        metadata, 
        name, 
        custom_image_url, 
        country_code, 
        opted_in_at, 
        created_at, 
        updated_at
    )
SELECT 
    
        tbl.id, 
        tbl.token_id, 
        tbl.owner_address, 
        tbl.mint_request_id, 
        tbl.user_id, 
        tbl.vin, 
        tbl.vin_confirmed, 
        tbl.claim_id,
        tbl.metadata, 
        tbl.name, 
        tbl.custom_image_url, 
        tbl.country_code, 
        tbl.opted_in_at, 
        tbl.created_at, 
        tbl.updated_at
    
FROM (
    SELECT 
        ud.id as id,
        vnft.token_id as token_id,
        vnft.owner_address as owner_address,
        vnft.mint_request_id as mint_request_id,
        ud.user_id as user_id,
        vnft.vin as vin,
        ud.vin_confirmed as vin_confirmed,
        vnft.claim_id as claim_id,
        ud.metadata as metadata,
        ud.name as name,
        ud.custom_image_url as custom_image_url,
        ud.country_code as country_code,
        ud.opted_in_at as opted_in_at,
        ud.created_at as created_at,
        ud.updated_at as updated_at
    FROM user_devices ud 
    LEFT JOIN vehicle_nfts vnft 
    ON ud.id = vnft.user_device_id
) tbl;

-- change contraints to align with new tbl
SET search_path = devices_api, public;
ALTER TABLE user_device_api_integrations DROP CONSTRAINT fkey_user_device_id;
ALTER TABLE user_device_api_integrations ADD CONSTRAINT fkey_user_device_id FOREIGN KEY (user_device_id) REFERENCES user_devices_v2(id) ON DELETE CASCADE;

ALTER TABLE user_device_data DROP CONSTRAINT fk_user_devices;
ALTER TABLE user_device_data ADD CONSTRAINT fk_user_devices FOREIGN KEY (user_device_id) REFERENCES user_devices_v2(id) ON DELETE CASCADE;

ALTER TABLE user_device_to_geofence DROP CONSTRAINT fk_user_device;
ALTER TABLE user_device_to_geofence ADD CONSTRAINT fk_user_device FOREIGN KEY (user_device_id) REFERENCES user_devices_v2(id) ON UPDATE CASCADE ON DELETE CASCADE;

ALTER TABLE autopi_jobs DROP CONSTRAINT fk_user_device;
ALTER TABLE autopi_jobs ADD CONSTRAINT fk_user_device FOREIGN KEY (user_device_id) REFERENCES user_devices_v2(id);

ALTER TABLE device_command_requests DROP CONSTRAINT device_command_requests_user_device_id_fkey;
ALTER TABLE device_command_requests ADD CONSTRAINT device_command_requests_user_device_id_fkey FOREIGN KEY (user_device_id) REFERENCES user_devices_v2(id) ON DELETE CASCADE;

ALTER TABLE aftermarket_devices DROP CONSTRAINT autopi_units_vehicle_token_id_fkey;
ALTER TABLE aftermarket_devices ADD CONSTRAINT autopi_units_vehicle_token_id_fkey FOREIGN KEY (vehicle_token_id) REFERENCES user_devices_v2(token_id) ON DELETE CASCADE;

ALTER TABLE error_code_queries DROP CONSTRAINT error_code_queries_user_device_id_fkey;
ALTER TABLE error_code_queries ADD CONSTRAINT error_code_queries_user_device_id_fkey FOREIGN KEY (user_device_id) REFERENCES user_devices_v2(id) ON DELETE CASCADE;

ALTER TABLE synthetic_devices DROP CONSTRAINT fkey_vehicle_token_id;
ALTER TABLE synthetic_devices ADD CONSTRAINT fkey_vehicle_token_id FOREIGN KEY (vehicle_token_id) REFERENCES user_devices_v2(token_id);

-- drop Original Tables

SET search_path = devices_api, public;
DROP TABLE vehicle_nfts;
DROP TABLE user_devices;



-- SELECT
--     constraint_name,
--     -- constraint_type,
--     table_name
-- FROM
--     information_schema.constraint_column_usage
-- WHERE
--     table_name IN ('user_devices', 'vehicle_nfts');

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';

SET search_path = devices_api, public;

CREATE TABLE vehicle_nfts(
    mint_request_id char(27)
        CONSTRAINT vehicle_nfts_mint_request_id_pkey PRIMARY KEY
        CONSTRAINT vehicle_nfts_mint_request_id_fkey REFERENCES meta_transaction_requests(id),
    user_device_id char(27)
        CONSTRAINT vehicle_nfts_user_device_id_fkey REFERENCES user_devices(id) ON DELETE SET NULL
        CONSTRAINT vehicle_nfts_user_device_id_key UNIQUE,
    vin char(17) NOT NULL, -- Want some constraint on minting the same VIN, but it's not clear what.
    token_id numeric(78, 0)
        CONSTRAINT vehicle_nfts_token_id_key UNIQUE,
    owner_address bytea
        CONSTRAINT vehicle_nfts_owner_address_check CHECK (length(owner_address) = 20),
    claim_id varchar
        CONSTRAINT vehicle_nfts_claim_id_key UNIQUE
        CONSTRAINT vehicle_nfts_claim_id_fkey REFERENCES verifiable_credentials(claim_id);
);

ALTER TABLE aftermarket_devices DROP CONSTRAINT autopi_units_vehicle_token_id_fkey;
ALTER TABLE aftermarket_devices ADD CONSTRAINT autopi_units_vehicle_token_id_fkey FOREIGN KEY (vehicle_token_id) REFERENCES vehicle_nfts(token_id) ON DELETE CASCADE;

CREATE TABLE user_devices
(
    id char(27) PRIMARY KEY,
    user_id text not null,
    device_definition_id char(27)    not null,
    vin_identifier text,
    name text, -- name the user can give
    custom_image_url text,
    country_code char(3),

    created_at           timestamptz not null default current_timestamp,
    updated_at           timestamptz not null default current_timestamp,

    PRIMARY KEY (id),
    CONSTRAINT fk_device_definition FOREIGN KEY (device_definition_id) REFERENCES device_definitions (id)
);

ALTER TABLE user_device_to_geofence DROP CONSTRAINT fk_user_device;
ALTER TABLE user_device_to_geofence ADD CONSTRAINT fk_user_device FOREIGN KEY (user_device_id) REFERENCES user_devices(id) ON UPDATE CASCADE ON DELETE CASCADE;

ALTER TABLE user_device_api_integrations DROP CONSTRAINT fkey_user_device_id;
ALTER TABLE user_device_api_integrations ADD CONSTRAINT fkey_user_device_id FOREIGN KEY (user_device_id) REFERENCES user_devices(id) ON DELETE CASCADE;

ALTER TABLE user_device_data DROP CONSTRAINT fk_user_devices;
ALTER TABLE user_device_data ADD CONSTRAINT fk_user_devices FOREIGN KEY (user_device_id) REFERENCES user_devices(id) ON DELETE CASCADE;

ALTER TABLE device_command_requests DROP CONSTRAINT device_command_requests_user_device_id_fkey;
ALTER TABLE device_command_requests ADD CONSTRAINT device_command_requests_user_device_id_fkey FOREIGN KEY (user_device_id) REFERENCES user_devices(id) ON DELETE CASCADE;

ALTER TABLE vehicle_nfts ADD CONSTRAINT vehicle_nfts_user_device_id_fkey FOREIGN KEY  (user_device_id) REFERENCES user_devices(id) ON DELETE SET NULL;

ALTER TABLE error_code_queries DROP CONSTRAINT error_code_queries_user_device_id_fkey;
ALTER TABLE error_code_queries ADD CONSTRAINT error_code_queries_user_device_id_fkey FOREIGN KEY (user_device_id) REFERENCES user_devices(id) ON DELETE CASCADE;

-- +goose StatementEnd
