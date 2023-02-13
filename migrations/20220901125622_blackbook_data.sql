-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';

SET search_path = devices_api, public;

ALTER TABLE external_vin_data add column blackbook_metadata jsonb;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';

SET search_path = devices_api, public;

ALTER TABLE external_vin_data drop column blackbook_metadata;

-- +goose StatementEnd
