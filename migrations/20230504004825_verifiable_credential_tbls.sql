-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
SET search_path = devices_api, public;

ALTER TABLE vehicle_nfts ADD COLUMN claim_id varchar;

CREATE TABLE verifiable_credentials
(
    claim_id varchar primary key,
    proof bytea not null
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
SET search_path = devices_api, public;

ALTER TABLE vehicle_nfts DROP COLUMN claim_id;
DROP TABLE verifiable_credentials;
-- +goose StatementEnd
