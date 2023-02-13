-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
SET search_path TO devices_api, public;
CREATE TABLE device_makes (
    id char(27) PRIMARY KEY, -- ksuid
    name text not null UNIQUE,
    external_ids jsonb, -- "data_provider": "id"
    created_at   timestamptz not null default current_timestamp,
    updated_at   timestamptz not null default current_timestamp
);
ALTER TABLE device_definitions ADD COLUMN device_make_id char(27); -- set fkey constraint
ALTER TABLE device_definitions
    ADD CONSTRAINT fk_device_make_id FOREIGN KEY (device_make_id) REFERENCES device_makes (id) on update cascade on delete restrict;
-- copy makes
insert into device_makes (id, name, external_ids)
SELECT DISTINCT ON (make) gen_random_ksuid(), make, jsonb_build_object(source, external_id) from device_definitions;
-- update DDs to set make_id
update device_definitions
set device_make_id=device_makes.id
from device_makes
where device_definitions.make=device_makes.name;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
SET search_path TO devices_api, public;
ALTER TABLE device_definitions DROP COLUMN device_make_id;
DROP TABLE device_makes;
-- +goose StatementEnd
