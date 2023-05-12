-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
SET search_path = devices_api, public;

ALTER TABLE vehicle_nfts ADD COLUMN claim_id varchar UNIQUE;

CREATE TABLE verifiable_credentials
(
    claim_id varchar 
        CONSTRAINT vehicle_nfts_credential_id_pkey PRIMARY KEY
        CONSTRAINT vehicle_nfts_credential_id_fkey REFERENCES vehicle_nfts(claim_id),
    "credential" bytea not null
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
SET search_path = devices_api, public;

DROP TABLE verifiable_credentials;
ALTER TABLE vehicle_nfts DROP COLUMN claim_id;
-- +goose StatementEnd
