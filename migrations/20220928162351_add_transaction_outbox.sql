-- +goose Up
-- +goose StatementBegin
SET search_path = devices_api, public;

CREATE TYPE meta_transaction_request_status AS ENUM ('Unsubmitted', 'Submitted', 'Mined', 'Confirmed');

CREATE TABLE meta_transaction_requests(
    id char(27) NOT NULL
        CONSTRAINT meta_transaction_requests_id_pkey PRIMARY KEY,
    status meta_transaction_request_status NOT NULL DEFAULT 'Unsubmitted',
    hash bytea
        CONSTRAINT meta_transaction_hash_key UNIQUE,
        CONSTRAINT meta_transaction_hash_check CHECK (length(hash) = 32),
    created_at timestamptz NOT NULL DEFAULT current_timestamp,
    updated_at timestamptz NOT NULL DEFAULT current_timestamp
);


ALTER TABLE user_devices ADD COLUMN mint_meta_transaction_request_id char(27)
    CONSTRAINT user_devices_mint_meta_transaction_request_id_key UNIQUE
    CONSTRAINT user_devices_mint_meta_transaction_request_id_fkey REFERENCES meta_transaction_requests(id);

ALTER TABLE autopi_units ADD COLUMN claim_meta_transaction_request_id char(27)
    CONSTRAINT autopi_units_claim_meta_transaction_request_id_key UNIQUE
    CONSTRAINT autopi_units_claim_meta_transaction_request_id_fkey REFERENCES meta_transaction_requests(id);

ALTER TABLE user_device_api_integrations ADD COLUMN pair_meta_transaction_request_id char(27)
    CONSTRAINT user_device_api_integrations_pair_meta_transaction_request_id_key UNIQUE
    CONSTRAINT user_device_api_integrations_pair_meta_transaction_request_id_fkey REFERENCES meta_transaction_requests(id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SET search_path = devices_api, public;

ALTER TABLE user_devices DROP COLUMN mint_request_id;
ALTER TABLE autopi_units DROP COLUMN claim_request_id;
ALTER TABLE user_device_api_integrations DROP COLUMN pair_request_id;

DROP TABLE meta_transaction_requests;
DROP TYPE meta_transaction_request_status;
-- +goose StatementEnd
