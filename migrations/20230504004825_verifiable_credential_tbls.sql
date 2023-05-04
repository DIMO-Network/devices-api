-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
SET search_path = devices_api, public;

CREATE TABLE verifiable_credentials
(
    user_device_id char(27) primary key,
    vin char(17) NOT NULL, 
    token_id numeric(78, 0),
    claims_root bytea,
    revocation_root bytea,
    root_of_roots bytea,
    "state" bytea,
    "id" bytea
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
SET search_path = devices_api, public;
DROP TABLE verifiable_credentials;
-- +goose StatementEnd
