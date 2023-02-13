-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';

SET search_path = devices_api, public;

alter table autopi_jobs
    drop constraint fk_user_device;

alter table autopi_jobs
    add constraint fk_user_device
        foreign key (user_device_id) references user_devices
            on update cascade on delete cascade;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';

SET search_path = devices_api, public;

alter table autopi_jobs
    drop constraint fk_user_device;

alter table autopi_jobs
    add constraint fk_user_device
        foreign key (user_device_id) references user_devices;

-- +goose StatementEnd
