-- +goose Up
-- +goose StatementBegin
SET search_path TO devices_api, public;

ALTER TABLE mint_requests
    ADD COLUMN vin char(17),
    ADD COLUMN child_device_id text,
    ALTER COLUMN user_device_id DROP NOT NULL,
    DROP CONSTRAINT mint_requests_user_device_id_fkey,
    ADD CONSTRAINT mint_requests_user_device_id_fkey FOREIGN KEY (user_device_id) REFERENCES user_devices(id) ON DELETE SET NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SET search_path TO devices_api, public;

ALTER TABLE mint_requests
    DROP COLUMN vin,
    DROP COLUMN child_device_id,
    ALTER COLUMN user_device_id SET NOT NULL,
    DROP CONSTRAINT mint_requests_user_device_id_fkey,
    ADD CONSTRAINT mint_requests_user_device_id_fkey FOREIGN KEY (user_device_id) REFERENCES user_devices(id) ON DELETE CASCADE;
-- +goose StatementEnd
