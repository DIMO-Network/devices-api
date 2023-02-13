-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
REVOKE CREATE ON schema public FROM public; -- public schema isolation
CREATE SCHEMA IF NOT EXISTS devices_api;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
DROP SCHEMA devices_api CASCADE;
GRANT CREATE, USAGE ON schema public TO public;
-- +goose StatementEnd