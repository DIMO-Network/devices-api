-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';

alter table devices_api.device_integrations
    drop constraint fk_device_definition;

alter table devices_api.device_integrations
    add constraint fk_device_definition
        foreign key (device_definition_id) references devices_api.device_definitions
            on delete cascade;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
SET search_path TO devices_api,public;
alter table devices_api.device_integrations
    drop constraint fk_device_definition;

alter table devices_api.device_integrations
    add constraint fk_device_definition
        foreign key (device_definition_id) references devices_api.device_definitions
-- +goose StatementEnd
