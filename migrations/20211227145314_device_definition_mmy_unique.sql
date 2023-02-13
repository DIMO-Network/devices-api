-- +goose Up
-- +goose StatementBegin
SET search_path = devices_api, public;
ALTER TABLE device_definitions ADD CONSTRAINT device_definitions_make_model_year_key UNIQUE (make, model, year);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SET search_path = devices_api, public;
ALTER TABLE device_definitions DROP CONSTRAINT device_definitions_make_model_year_key;
-- +goose StatementEnd
