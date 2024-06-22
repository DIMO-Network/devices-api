-- +goose Up
-- +goose StatementBegin
SET search_path = devices_api, public;

CREATE OR REPLACE VIEW vehicle_nfts AS
    SELECT 
            ud.id as user_device_id,
            ud.mint_request_id,
            ud.token_id,
            NULL::varchar as claim_id, -- Used to be the claim id.
            ud.owner_address,
            ud.vin_identifier as vin
    FROM user_devices ud;

ALTER TABLE user_devices DROP COLUMN claim_id;

DROP TABLE verifiable_credentials;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SET search_path = devices_api, public;
-- +goose StatementEnd
