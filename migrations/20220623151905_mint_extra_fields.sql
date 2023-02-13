-- +goose Up
-- +goose StatementBegin
SET search_path = devices_api, public;

ALTER TABLE mint_requests ADD COLUMN tx_hash bytea;
ALTER TABLE mint_requests
    DROP CONSTRAINT mint_requests_user_device_id_fkey,
    ADD CONSTRAINT mint_requests_user_device_id_fkey FOREIGN KEY (user_device_id) REFERENCES user_devices(id) ON DELETE CASCADE;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SET search_path = devices_api, public;

ALTER TABLE mint_requests DROP COLUMN tx_hash;
-- +goose StatementEnd
