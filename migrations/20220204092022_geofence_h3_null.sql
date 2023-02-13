-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
SET search_path TO devices_api,public;
ALTER TABLE geofences ALTER COLUMN h3_indexes DROP NOT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
SET search_path TO devices_api,public;
ALTER TABLE geofences ALTER COLUMN h3_indexes SET NOT NULL;
-- +goose StatementEnd
