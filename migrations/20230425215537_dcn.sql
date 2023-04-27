-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
SET search_path = devices_api, public;
create table dcn
(
    nft_node_id   bytea
        constraint dcn_nft_node_address_check
            check (length(nft_node_id) = 32),
    owner_address   bytea
        constraint dcn_owner_address_check
            check (length(owner_address) = 20),
    name text unique,
    expiration timestamp with time zone,
    nft_node_block_create_time timestamp with time zone,
    created_at           timestamp with time zone default CURRENT_TIMESTAMP not null,
    updated_at           timestamp with time zone default CURRENT_TIMESTAMP not null,
    primary key (nft_node_id)
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
SET search_path = devices_api, public;
drop table dcn;

-- +goose StatementEnd
