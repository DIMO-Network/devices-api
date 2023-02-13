-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
SET search_path = devices_api, public;

alter table autopi_jobs add column command_result jsonb;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
SET search_path = devices_api, public;

alter table autopi_jobs drop column command_result;
-- +goose StatementEnd
