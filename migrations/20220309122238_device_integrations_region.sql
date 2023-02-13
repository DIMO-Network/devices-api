-- +goose Up
-- +goose StatementBegin
SET search_path TO devices_api, public;

-- Existing integrations are for USA or USA, CAN, UMI; these are all rolled up into Americas.
DELETE FROM device_integrations WHERE country != 'USA';

ALTER TABLE device_integrations ADD COLUMN region text;
UPDATE device_integrations SET region = 'Americas';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SET search_path TO devices_api, public;

-- We could recreate the 3 for each record here.
ALTER TABLE device_integrations DROP COLUMN region;
-- +goose StatementEnd
