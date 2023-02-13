-- +goose Up
-- +goose StatementBegin
SET search_path TO devices_api, public;

ALTER TABLE device_makes ADD COLUMN token_id numeric(78, 0);

ALTER TABLE device_makes ADD CONSTRAINT device_makes_token_id_key UNIQUE (token_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SET search_path TO devices_api, public;

ALTER TABLE device_makes DROP COLUMN token_id;
-- +goose StatementEnd
