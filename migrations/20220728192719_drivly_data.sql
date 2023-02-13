-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
-- +goose StatementEnd


CREATE TABLE drivly_data
(
    id                   char(27), -- ksuid
    device_definition_id char(27), -- ksuid
    VIN                  text        not null,
    user_device_id       char(27), -- ksuid
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
    created_at           timestamptz not null default current_timestamp,
    updated_at           timestamptz not null default current_timestamp,
    PRIMARY KEY (id),
    CONSTRAINT fk_device_definition FOREIGN KEY (device_definition_id) REFERENCES device_definitions (id) ON DELETE CASCADE,
    CONSTRAINT fk_user_device       FOREIGN KEY (user_device_id) REFERENCES user_devices (id) ON DELETE SET NULL
);

ALTER TABLE user_devices
    ADD COLUMN device_style_id char(27);

ALTER TABLE user_devices
    ADD CONSTRAINT user_devices_styles FOREIGN KEY (device_style_id) REFERENCES device_styles (id);

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd

ALTER TABLE user_devices
    DROP COLUMN device_style_id;

DROP TABLE drivly_data;
