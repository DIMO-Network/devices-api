-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
SET search_path = devices_api, public;
CREATE TABLE user_device_data
(
    user_device_id char(27) PRIMARY KEY, -- ksuid
    data           jsonb,
    created_at     timestamptz not null default current_timestamp,
    updated_at     timestamptz not null default current_timestamp,
    CONSTRAINT fk_user_devices
        FOREIGN KEY(user_device_id)
            REFERENCES user_devices(id) ON DELETE CASCADE
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
SET search_path = devices_api, public;
DROP TABLE user_device_data;
-- +goose StatementEnd
