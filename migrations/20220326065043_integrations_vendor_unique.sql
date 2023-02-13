-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
SET search_path = devices_api, public;
ALTER TABLE integrations DROP CONSTRAINT idx_integrations_type_style_vendor;
ALTER TABLE integrations ADD CONSTRAINT idx_integrations_vendor UNIQUE (vendor);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
SET search_path = devices_api, public;
ALTER TABLE integrations DROP CONSTRAINT idx_integrations_vendor;
ALTER TABLE integrations ADD CONSTRAINT idx_integrations_type_style_vendor UNIQUE (type, style, vendor);
-- +goose StatementEnd
