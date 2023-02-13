-- +goose Up
-- +goose StatementBegin
SET search_path TO devices_api, public;

ALTER TABLE device_definitions DROP COLUMN make;
ALTER TABLE device_integrations DROP COLUMN country;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SET search_path TO devices_api, public;

ALTER TABLE device_definitions ADD COLUMN make text not null default 'unknown';
ALTER TABLE device_integrations ADD COLUMN country text not null default 'USA';
-- +goose StatementEnd
