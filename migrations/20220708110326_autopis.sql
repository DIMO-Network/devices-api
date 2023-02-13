-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
SET search_path = devices_api, public;

CREATE TABLE autopi_units
(
    autopi_unit_id         char(36) PRIMARY KEY,
    autopi_device_id       char(36),
    user_id         text not null,
    nft_address     varchar(100),
    created_at      timestamptz not null default current_timestamp,
    updated_at      timestamptz not null default current_timestamp
);
CREATE UNIQUE INDEX autopi_units_nft_address_idx ON autopi_units (nft_address);
CREATE UNIQUE INDEX autopi_units_device_id_idx ON autopi_units (autopi_device_id);

ALTER TABLE user_device_api_integrations
ADD COLUMN autopi_unit_id char(36);

ALTER TABLE user_device_api_integrations ADD CONSTRAINT user_device_api_integrations_autopi_units FOREIGN KEY (autopi_unit_id) REFERENCES autopi_units(autopi_unit_id);

ALTER TABLE autopi_jobs
ADD COLUMN autopi_unit_id char(36);

ALTER TABLE autopi_jobs ADD CONSTRAINT autopi_jobs_autopi_units FOREIGN KEY (autopi_unit_id) REFERENCES autopi_units(autopi_unit_id);

INSERT INTO autopi_units (autopi_unit_id, autopi_device_id, user_id)
SELECT
  json_extract_path_text(udai.metadata::json,'autoPiUnitId') unit_id,
  udai.external_id,
  ud.user_id
FROM
    user_device_api_integrations udai
    INNER JOIN integrations i ON ( udai.integration_id = i.id ) 
    INNER JOIN user_devices ud ON ( udai.user_device_id = ud.id )
WHERE
    i.vendor = 'AutoPi';

UPDATE
    user_device_api_integrations udai
SET
    autopi_unit_id = a.autopi_unit_id
FROM
    autopi_units a
WHERE
    a.autopi_unit_id = json_extract_path_text(udai.metadata::json,'autoPiUnitId');

UPDATE
    autopi_jobs aj
SET
    autopi_unit_id = a.autopi_unit_id
FROM
    autopi_units a
WHERE
    a.autopi_device_id = aj.autopi_device_id;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';

ALTER TABLE autopi_jobs drop column autopi_unit_id;
ALTER TABLE user_device_api_integrations DROP COLUMN autopi_unit_id;

DROP TABLE autopi_units;

-- +goose StatementEnd
