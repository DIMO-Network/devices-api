-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';

ALTER TABLE vehicle_nfts ADD COLUMN burn_request_id CHAR(27);

ALTER TABLE vehicle_nfts ADD CONSTRAINT vehicle_nfts_burn_request_id_key UNIQUE (burn_request_id);
ALTER TABLE vehicle_nfts ADD CONSTRAINT vehicle_nfts_burn_request_id_fkey FOREIGN KEY (burn_request_id) REFERENCES meta_transaction_requests(id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';

ALTER TABLE vehicle_nfts DROP CONSTRAINT vehicle_nfts_burn_request_id_fkey;
ALTER TABLE vehicle_nfts DROP CONSTRAINT vehicle_nfts_burn_request_id_key;

ALTER TABLE vehicle_nfts DROP COLUMN burn_request_id;
-- +goose StatementEnd
