-- +goose Up
-- +goose StatementBegin
SET search_path TO devices_api, public;

ALTER TABLE device_integrations ALTER COLUMN region SET NOT NULL;

ALTER TABLE device_integrations DROP CONSTRAINT device_integrations_pkey;
ALTER TABLE device_integrations ADD CONSTRAINT pkey_device_region PRIMARY KEY (device_definition_id, integration_id, region);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SET search_path TO devices_api, public;

ALTER TABLE device_integrations DROP CONSTRAINT pkey_device_region;
ALTER TABLE device_integrations ADD CONSTRAINT device_integrations_pkey PRIMARY KEY (device_definition_id, integration_id, country);

ALTER TABLE device_integrations ALTER COLUMN region DROP NOT NULL;
-- +goose StatementEnd
