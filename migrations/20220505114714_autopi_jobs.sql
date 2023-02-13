-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
SET search_path = devices_api, public;

create table autopi_jobs
(
    id                   varchar(50) not null -- autopi job id
        constraint autopi_jobs_pk
            primary key,
    autopi_device_id     text        not null,
    command              text        not null,
    state                varchar(100)         default 'Sent' not null,
    command_last_updated timestamptz,
    user_device_id       char(27),            -- if command run on a user_device
    created_at           timestamptz not null default current_timestamp,
    updated_at           timestamptz not null default current_timestamp,

    CONSTRAINT fk_user_device FOREIGN KEY (user_device_id) REFERENCES user_devices (id)
);
comment on table autopi_jobs is 'used to track commands sent to an auotpi device';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
SET search_path = devices_api, public;

drop table autopi_jobs;
-- +goose StatementEnd
