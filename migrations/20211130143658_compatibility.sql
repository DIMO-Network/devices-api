-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
SET search_path = devices_api, public;

CREATE TYPE integration_type AS ENUM (
    'Hardware',
    'API'
    );
CREATE TYPE integration_style AS ENUM (
    'Addon',
    'OEM',
    'Webhook'
    );

CREATE TABLE integrations
(
    id         char(27) PRIMARY KEY, -- ksuid
    type       integration_type  not null,
    style      integration_style not null,
    vendor    varchar(50)       not null,

    created_at timestamptz       not null default current_timestamp,
    updated_at timestamptz       not null default current_timestamp
);

CREATE TABLE device_integrations
(
    device_definition_id char(27)    not null,
    integration_id       char(27)    not null,
    country              char(3),
    capabilities         jsonb,

    created_at           timestamptz not null default current_timestamp,
    updated_at           timestamptz not null default current_timestamp,

    PRIMARY KEY (device_definition_id, integration_id, country),
    CONSTRAINT fk_device_definition FOREIGN KEY (device_definition_id) REFERENCES device_definitions (id),
    CONSTRAINT fk_integration FOREIGN KEY (integration_id) REFERENCES integrations (id)
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
SET search_path = devices_api, public;

DROP TABLE device_integrations;
DROP TABLE integrations;
DROP TYPE integration_style;
DROP TYPE integration_type;
-- +goose StatementEnd
