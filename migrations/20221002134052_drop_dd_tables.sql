-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
SET search_path = devices_api, public;

alter table user_devices
    drop constraint user_devices_styles;
alter table user_devices
    drop constraint fk_device_definition;
alter table external_vin_data
    drop constraint fk_device_definition;
alter table user_device_api_integrations
    drop constraint fkey_integration_id;
alter table device_command_requests
    drop constraint device_command_requests_integration_id_fkey;

drop table device_integrations;
drop table device_styles;
drop table device_definitions;
drop table device_makes;
drop table integrations cascade;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
SET search_path = devices_api, public;

-- auto-generated definition
create table integrations
(
    id                 char(27)                                           not null
        primary key,
    type               device_definitions_api.integration_type            not null,
    style              device_definitions_api.integration_style           not null,
    vendor             varchar(50)                                        not null
        constraint idx_integrations_vendor
            unique,
    created_at         timestamp with time zone default CURRENT_TIMESTAMP not null,
    updated_at         timestamp with time zone default CURRENT_TIMESTAMP not null,
    refresh_limit_secs integer                  default 3600              not null,
    metadata           jsonb
);

comment on column integrations.refresh_limit_secs is 'How often can integration be called in seconds';

-- auto-generated definition
create table device_makes
(
    id                char(27)                                           not null
        primary key,
    name              text                                               not null
        unique,
    external_ids      jsonb,
    created_at        timestamp with time zone default CURRENT_TIMESTAMP not null,
    updated_at        timestamp with time zone default CURRENT_TIMESTAMP not null,
    token_id          numeric(78)
        unique,
    logo_url          text,
    oem_platform_name text
);

-- auto-generated definition
create table device_definitions
(
    id             char(27)                                           not null
        primary key,
    model          varchar(100)                                       not null,
    year           smallint                                           not null,
    image_url      text,
    metadata       jsonb,
    created_at     timestamp with time zone default CURRENT_TIMESTAMP not null,
    updated_at     timestamp with time zone default CURRENT_TIMESTAMP not null,
    source         text,
    verified       boolean                  default false             not null,
    external_id    text,
    device_make_id char(27)                                           not null
        constraint fk_device_make_id
            references device_makes
            on update cascade on delete restrict,
    constraint idx_device_make_id_model_year
        unique (device_make_id, model, year)
);

-- auto-generated definition
create table device_integrations
(
    device_definition_id char(27)                                           not null
        constraint fk_device_definition
            references device_definitions
            on delete cascade,
    integration_id       char(27)                                           not null
        constraint fk_integration
            references integrations,
    capabilities         jsonb,
    created_at           timestamp with time zone default CURRENT_TIMESTAMP not null,
    updated_at           timestamp with time zone default CURRENT_TIMESTAMP not null,
    region               text                                               not null,
    constraint pkey_device_region
        primary key (device_definition_id, integration_id, region)
);

-- auto-generated definition
create table device_styles
(
    id                   char(27)                                           not null
        primary key,
    device_definition_id char(27)                                           not null
        constraint fk_device_definition
            references device_definitions,
    name                 text                                               not null,
    external_style_id    text                                               not null,
    source               text                                               not null,
    created_at           timestamp with time zone default CURRENT_TIMESTAMP not null,
    updated_at           timestamp with time zone default CURRENT_TIMESTAMP not null,
    sub_model            text                                               not null
);

create unique index device_definition_name_sub_modelx
    on device_styles (device_definition_id, source, name, sub_model);

create unique index device_definition_style_idx
    on device_styles (device_definition_id, source, external_style_id);

-- +goose StatementEnd
