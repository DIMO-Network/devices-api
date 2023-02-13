-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
SET search_path = devices_api, public;
ALTER TABLE integrations ADD COLUMN refresh_limit_secs int not null default 3600;
COMMENT ON COLUMN integrations.refresh_limit_secs IS 'How often can integration be called in seconds';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
SET search_path = devices_api, public;
ALTER TABLE integrations DROP COLUMN refresh_limit_secs;
-- +goose StatementEnd
