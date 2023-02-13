-- +goose Up
-- +goose StatementBegin
SET search_path = devices_api, public;

CREATE TABLE nft_privileges(
    contract_address bytea NOT NULL
        CONSTRAINT nft_privileges_contract_address_check CHECK (length(contract_address) = 20),
    token_id numeric(78, 0) NOT NULL,
    privilege bigint not null,
    user_address bytea NOT NULL
        CONSTRAINT nft_privileges_user_address_check CHECK (length(user_address) = 20),
    expiry timestamptz not null,
    
    created_at           timestamptz not null default current_timestamp,
    updated_at           timestamptz not null default current_timestamp,

    CONSTRAINT nft_privileges_pkey PRIMARY KEY (contract_address, token_id, privilege, user_address)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SET search_path = devices_api, public;

DROP TABLE nft_privileges;
-- +goose StatementEnd
