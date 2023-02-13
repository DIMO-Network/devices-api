-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
SET search_path TO devices_api,public;
CREATE TYPE geofence_type AS ENUM (
    'PrivacyFence',
    'TriggerEntry',
    'TriggerExit'
    );

CREATE TABLE geofences (
    id char(27) PRIMARY KEY, -- ksuid
    user_id text not null, -- user_id that owns it
    name text not null,
    type geofence_type not null default 'PrivacyFence',
    h3_indexes text[] not null,
    created_at   timestamptz not null default current_timestamp,
    updated_at   timestamptz not null default current_timestamp
);
CREATE UNIQUE INDEX user_id_name on geofences (user_id, name);

CREATE TABLE user_device_to_geofence (
    user_device_id char(27), -- ksuid
    geofence_id char(27),
    created_at   timestamptz not null default current_timestamp,
    updated_at   timestamptz not null default current_timestamp,
    PRIMARY KEY (user_device_id, geofence_id),
    CONSTRAINT fk_user_device FOREIGN KEY (user_device_id) REFERENCES user_devices (id) on update cascade on delete cascade,
    CONSTRAINT fk_geofence FOREIGN KEY (geofence_id) REFERENCES geofences (id) on update cascade on delete cascade
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
SET search_path TO devices_api,public;
DROP TABLE user_device_to_geofence;
DROP TABLE geofences;
DROP TYPE geofence_type;
-- +goose StatementEnd
