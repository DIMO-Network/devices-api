-- +goose Up
-- +goose StatementBegin
SET search_path = devices_api, public;

CREATE TABLE verifiable_credentials(
    claim_id varchar CONSTRAINT verifiable_credentials_pkey PRIMARY KEY,
    "credential" bytea NOT NULL
);

ALTER TABLE vehicle_nfts ADD COLUMN claim_id varchar
    CONSTRAINT vehicle_nfts_claim_id_key UNIQUE
    CONSTRAINT vehicle_nfts_claim_id_fkey REFERENCES verifiable_credentials(claim_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SET search_path = devices_api, public;

DROP TABLE verifiable_credentials;
ALTER TABLE vehicle_nfts DROP COLUMN claim_id;
-- +goose StatementEnd
