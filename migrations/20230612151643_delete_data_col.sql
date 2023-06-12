-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';

SET search_path = devices_api, public;
alter table user_device_data drop column data;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
SET search_path = devices_api, public;

alter table user_device_data add column data jsonb;
-- +goose StatementEnd
