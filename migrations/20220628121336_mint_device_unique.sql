-- +goose Up
-- +goose StatementBegin
SET search_path = devices_api, public;

ALTER TABLE mint_requests ADD CONSTRAINT mint_requests_user_device_id_key UNIQUE (user_device_id);

ALTER TABLE user_devices DROP COLUMN token_id;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SET search_path = devices_api, public;

ALTER TABLE mint_requests DROP CONSTRAINT mint_requests_user_device_id_key;

-- Copy over the correct values from mint_requests.
ALTER TABLE user_devices ADD COLUMN token_id numeric(78, 0);
ALTER TABLE user_devices ADD CONSTRAINT user_devices_token_id_key UNIQUE (token_id);

UPDATE user_devices ud SET token_id = mr.token_id FROM mint_requests mr WHERE mr.user_device_id = ud.id;

ALTER TABLE mint_requests DROP CONSTRAINT mint_requests_user_device_id_key;
-- +goose StatementEnd
