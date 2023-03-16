-- +goose Up
-- +goose StatementBegin
CREATE TABLE devices_api.error_code_queries(
    id char(27) not null,
    user_device_id char(27)    not null,
    error_codes text[] not null,
    query_response     text not null,
    
    created_at           timestamptz not null default current_timestamp,
    updated_at           timestamptz not null default current_timestamp,

    PRIMARY KEY (id),
    CONSTRAINT error_code_queries_user_device_id_fkey FOREIGN KEY (user_device_id) REFERENCES user_devices(id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE devices_api.error_code_quries
-- +goose StatementEnd
