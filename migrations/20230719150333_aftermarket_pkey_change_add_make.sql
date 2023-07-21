-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
SET search_path = devices_api, public;
alter table aftermarket_devices add column device_manufacturer_token_id numeric(78)
    CHECK (device_manufacturer_token_id >= 0 AND device_manufacturer_token_id < 2^256);

-- delete data that won't work for this phase, address cannot be null
delete from autopi_jobs where autopi_unit_id in (
    select serial from aftermarket_devices where ethereum_address is null
    );
delete from user_device_api_integrations where integration_id = '27qftVRWQYpVDcO5DltO5Ojbjxk'
                                           and user_device_id in (
        select user_device_id from user_device_api_integrations inner join aftermarket_devices ad on ad.serial = user_device_api_integrations.serial
        where ad.ethereum_address is null
    );
delete from aftermarket_devices where ethereum_address is null;
alter table aftermarket_devices alter column ethereum_address set not null;

-- drop fkey relations
ALTER TABLE user_device_api_integrations
    DROP CONSTRAINT user_device_api_integrations_autopi_units;

ALTER TABLE autopi_jobs
    DROP CONSTRAINT autopi_jobs_autopi_units;

-- drop the pkey and recreate index on serial
alter table aftermarket_devices drop constraint autopi_units_pkey;
create unique index aftermarket_devices_serial_idx on aftermarket_devices (serial) ;

ALTER TABLE aftermarket_devices ADD PRIMARY KEY (ethereum_address);

-- recreate fkey relations, but keep relation to serial since we use it as the autopi unit id for other operations
alter table user_device_api_integrations
    add constraint user_device_api_integrations_aftermarket_devices
        foreign key (serial) references aftermarket_devices (serial);
alter table autopi_jobs
    add constraint autopi_jobs_aftermarket_devices
        foreign key (autopi_unit_id) references aftermarket_devices (serial);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
SET search_path = devices_api, public;
alter table aftermarket_devices drop column device_make_token_id;
alter table aftermarket_devices alter column ethereum_address drop not null;
alter table aftermarket_devices drop constraint aftermarket_devices_pkey;
ALTER TABLE aftermarket_devices ADD PRIMARY KEY (serial);

-- +goose StatementEnd