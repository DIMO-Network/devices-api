-- +goose Up
-- +goose StatementBegin
SET search_path = devices_api, public;
ALTER TABLE user_devices ADD COLUMN definition_id text not null default '';
-- +goose StatementEnd


-- +goose Down
-- +goose StatementBegin
SET search_path = devices_api, public;
ALTER TABLE user_devices DROP COLUMN definition_id;
-- +goose StatementEnd
