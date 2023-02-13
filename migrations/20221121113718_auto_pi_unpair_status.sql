-- +goose Up
-- +goose StatementBegin
SET search_path = devices_api, public;

ALTER TABLE user_device_api_integrations
    ADD COLUMN unpair_meta_transaction_request_id char(27)
        CONSTRAINT user_device_api_integrations_unpair_transaction_request_id_key UNIQUE
        CONSTRAINT user_device_api_integrations_unpair_transaction_request_id_fkey REFERENCES meta_transaction_requests(id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SET search_path = devices_api, public;

ALTER TABLE user_device_api_integrations DROP COLUMN unpair_meta_transaction_request_id;
-- +goose StatementEnd
