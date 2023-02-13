-- +goose Up
-- +goose StatementBegin
SET search_path = devices_api, public;
UPDATE device_definitions SET model = TRIM(BOTH FROM model); -- Might hit the uniqueness constraint if the DB is odd
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- Impossible to say
-- +goose StatementEnd
