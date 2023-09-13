-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
SET search_path = devices_api, public;

drop table external_vin_data;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
SET search_path = devices_api, public;

create table external_vin_data
(
    id                   char(27)                                           not null
        constraint drivly_data_pkey
            primary key,
    device_definition_id char(27),
    vin                  text                                               not null,
    user_device_id       char(27),
    vin_metadata         json,
    offer_metadata       json,
    autocheck_metadata   json,
    build_metadata       json,
    cargurus_metadata    json,
    carvana_metadata     json,
    carmax_metadata      json,
    carstory_metadata    json,
    edmunds_metadata     json,
    tmv_metadata         json,
    kbb_metadata         json,
    vroom_metadata       json,
    created_at           timestamp with time zone default CURRENT_TIMESTAMP not null,
    updated_at           timestamp with time zone default CURRENT_TIMESTAMP not null,
    pricing_metadata     jsonb,
    blackbook_metadata   jsonb,
    request_metadata     jsonb,
    vincario_metadata    jsonb
);

-- +goose StatementEnd
