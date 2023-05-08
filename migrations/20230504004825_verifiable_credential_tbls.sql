-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
SET search_path = devices_api, public;

CREATE TABLE verifiable_credentials
(
    token_id char(27) primary key,
    x bytea not null,
    y bytea not null,
    d bytea not null,
    identity varchar not null
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
SET search_path = devices_api, public;
DROP TABLE verifiable_credentials;
-- +goose StatementEnd
