-- +goose Up
-- +goose StatementBegin
SET search_path = devices_api, public;

DROP TABLE mint_requests;

DROP TYPE txstate;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SET search_path = devices_api, public;

-- Too much to do.

-- +goose StatementEnd
