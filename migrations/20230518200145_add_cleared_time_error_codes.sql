-- +goose Up
-- +goose StatementBegin
SET search_path = devices_api, public;

ALTER TABLE devices_api.error_code_queries ADD column cleared_at timestamptz null;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SET search_path = devices_api, public;

ALTER TABLE devices_api.error_code_queries DROP COLUMN cleared_at;
-- +goose StatementEnd
