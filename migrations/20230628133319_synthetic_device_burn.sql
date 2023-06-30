-- +goose Up
-- +goose StatementBegin
SET search_path = devices_api, public;

ALTER TABLE synthetic_devices ADD COLUMN burn_request_id varchar;

ALTER TABLE synthetic_devices ADD CONSTRAINT synthetic_devices_burn_request_id_key UNIQUE (burn_request_id);
ALTER TABLE synthetic_devices ADD CONSTRAINT synthetic_devices_burn_request_id_fkey FOREIGN KEY (burn_request_id) REFERENCES meta_transaction_requests(id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SET search_path = devices_api, public;

ALTER TABLE synthetic_devices DROP COLUMN burn_request_id;
-- +goose StatementEnd
