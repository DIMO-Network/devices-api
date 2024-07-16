-- +goose Up
-- +goose StatementBegin
DROP VIEW vehicle_nfts;

CREATE VIEW vehicle_nfts AS
SELECT
    ud.id AS user_device_id,
    ud.mint_request_id,
    ud.token_id,
    NULL::text AS claim_id,
    ud.owner_address,
    ud.vin_identifier AS vin
FROM
    user_devices ud
WHERE
    ud.mint_request_id IS NOT NULL;

ALTER TABLE user_devices DROP COLUMN claim_id;

DROP TABLE verifiable_credentials;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
