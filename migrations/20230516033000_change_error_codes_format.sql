-- +goose Up
-- +goose StatementBegin
SET search_path = devices_api, public;
ALTER TABLE devices_api.error_code_queries DROP COLUMN query_response;
ALTER TABLE devices_api.error_code_queries DROP COLUMN error_codes;
ALTER TABLE devices_api.error_code_queries ADD column codes_query_response jsonb;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SET search_path = devices_api, public;

ALTER TABLE devices_api.error_code_queries DROP COLUMN codes_query_response;
ALTER TABLE devices_api.error_code_queries ADD column error_codes text[] not null;
ALTER TABLE devices_api.error_code_queries ADD column query_response text not null;
-- +goose StatementEnd
