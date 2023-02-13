-- +goose Up
-- +goose StatementBegin
SET search_path = devices_api, public;

ALTER TABLE autopi_units
    ALTER COLUMN user_id DROP NOT NULL,
    ALTER COLUMN nft_address TYPE bytea USING decode(substring(nft_address from 2), 'hex'),
    ADD COLUMN token_id numeric(78, 0),
    ADD CONSTRAINT autopi_units_token_id_key UNIQUE (token_id);

ALTER TABLE autopi_units RENAME nft_address to ethereum_address;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SET search_path = devices_api, public;

ALTER TABLE autopi_units
    ALTER COLUMN user_id SET NOT NULL,
    ALTER COLUMN nft_address TYPE text,
    DROP COLUMN token_id;
-- +goose StatementEnd
