-- +goose Up
-- +goose StatementBegin
SET search_path = devices_api, public;

ALTER TABLE synthetic_devices DROP CONSTRAINT synthetic_devices_pkey;
ALTER TABLE synthetic_devices ALTER COLUMN vehicle_token_id DROP NOT NULL;
ALTER TABLE synthetic_devices ADD CONSTRAINT synthetic_devices_vehicle_token_id_key UNIQUE (vehicle_token_id);
ALTER TABLE synthetic_devices ADD CONSTRAINT synthetic_devices_pkey PRIMARY KEY (mint_request_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SET search_path = devices_api, public;
-- +goose StatementEnd
