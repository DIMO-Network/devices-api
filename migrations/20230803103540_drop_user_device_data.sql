-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
SET search_path = devices_api, public;
drop table user_device_data;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
SET search_path = devices_api, public;
create table user_device_data
(
    user_device_id              char(27)                                           not null
        constraint fk_user_devices
            references user_devices
            on delete cascade,
    created_at                  timestamp with time zone default CURRENT_TIMESTAMP not null,
    updated_at                  timestamp with time zone default CURRENT_TIMESTAMP not null,
    error_data                  jsonb,
    last_odometer_event_at      timestamp with time zone,
    integration_id              char(27)                                           not null,
    real_last_odometer_event_at timestamp with time zone,
    signals                     jsonb,
    primary key (user_device_id, integration_id)
);
-- +goose StatementEnd
