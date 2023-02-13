-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
SET search_path = devices_api, public;
-- map all integration_id's via their respective user_device_api_integration
update user_device_data udd set integration_id =
                                    (select udai.integration_id from user_device_api_integrations udai where udd.user_device_id = udai.user_device_id
                                     order by udd.created_at limit 1)
where udd.integration_id is null;
-- any left over assume they are smartcar should be safe
update user_device_data set integration_id = (select id from integrations where vendor = 'SmartCar') where integration_id is null;

alter table user_device_data
    drop constraint user_device_data_pkey;

alter table user_device_data
    add primary key (user_device_id, integration_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
SET search_path = devices_api, public;

alter table user_device_data
    drop constraint user_device_data_pkey;

alter table user_device_data
    add primary key (user_device_id);

-- +goose StatementEnd
