-- +goose Up
-- +goose StatementBegin
ALTER TABLE verifiable_credentials
    ADD COLUMN credential_jsonb JSONB;

UPDATE verifiable_credentials
    SET credential_jsonb = encode(credential, 'escape')::jsonb;

ALTER TABLE verifiable_credentials
    DROP COLUMN "credential";

ALTER TABLE verifiable_credentials
    RENAME COLUMN credential_jsonb TO "credential";

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE verifiable_credentials
    ADD COLUMN credential_bytea BYTEA;

UPDATE verifiable_credentials
    SET credential_bytea = convert_to(credential::TEXT, 'LATIN1');

ALTER TABLE verifiable_credentials
    DROP COLUMN "credential";

ALTER TABLE verifiable_credentials
    RENAME COLUMN credential_bytea TO "credential";

-- +goose StatementEnd
