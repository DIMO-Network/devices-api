-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
SET search_path TO devices_api,public;

ALTER TABLE device_definitions DROP COLUMN sub_model;
ALTER TABLE device_definitions ADD COLUMN sub_models text[];
ALTER TABLE device_definitions ADD COLUMN external_id text;

CREATE TABLE device_styles
(
    id char(27) PRIMARY KEY, -- ksuid
    device_definition_id char(27) NOT NULL,
    name TEXT NOT NULL,
    external_style_id TEXT NOT NULL, -- eg. edmunds styleId
    source TEXT NOT NULL, -- eg. edmunds or api1, some form of identifyin the api source
    created_at   timestamptz not null default current_timestamp,
    updated_at   timestamptz not null default current_timestamp,
    CONSTRAINT fk_device_definition FOREIGN KEY (device_definition_id) REFERENCES device_definitions (id)
);
CREATE UNIQUE INDEX device_definition_style_idx ON device_styles (device_definition_id, source,  external_style_id);
CREATE UNIQUE INDEX device_definition_namex ON device_styles (device_definition_id, source,  name);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
SET search_path TO devices_api,public;
DROP TABLE device_styles;
ALTER TABLE device_definitions DROP COLUMN external_id;
ALTER TABLE device_definitions DROP COLUMN sub_models; -- currently sub_model is not being used so won't re-create


-- +goose StatementEnd
