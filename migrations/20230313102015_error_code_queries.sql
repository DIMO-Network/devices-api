-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';

CREATE TABLE devices_api.error_code_queries(
    id char(27) not null,
    user_device_id char(27)    not null,
    error_codes text[] not null,
    query_response     text not null,
    
    created_at           timestamptz not null default current_timestamp,
    updated_at           timestamptz not null default current_timestamp,

    PRIMARY KEY (id),
    CONSTRAINT fkey_user_device_id FOREIGN KEY (user_device_id) REFERENCES user_devices(id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';

DROP TABLE devices_api.error_code_quries
-- +goose StatementEnd
