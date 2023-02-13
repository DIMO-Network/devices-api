-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
SET search_path = devices_api, public;

ALTER TABLE user_device_data ADD COLUMN integration_id char(27);

alter table user_device_data
    add constraint user_device_data_integrations_id_fk
        foreign key (integration_id) references integrations;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
SET search_path = devices_api, public;

ALTER TABLE user_device_data DROP COLUMN integration_id;
-- +goose StatementEnd
