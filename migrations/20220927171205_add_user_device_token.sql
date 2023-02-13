-- +goose Up
-- +goose StatementBegin
SET search_path = devices_api, public;

ALTER TABLE user_devices
    ADD COLUMN token_id numeric(78, 0),
    ADD CONSTRAINT user_devices_token_id_key UNIQUE (token_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SET search_path = devices_api, public;

ALTER TABLE user_devices
    DROP COLUMN token_id;
-- +goose StatementEnd
