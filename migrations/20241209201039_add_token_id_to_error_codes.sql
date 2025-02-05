-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
SET search_path = devices_api, public;
ALTER TABLE error_code_queries
    ADD vehicle_token_id NUMERIC(78, 0);

UPDATE error_code_queries ecq SET vehicle_token_id = ud.token_id FROM user_devices ud WHERE ud.id = ecq.user_device_id;

-- ALTER TABLE error_code_queries
--     ALTER COLUMN user_device_token_id SET NOT NULL;

ALTER TABLE error_code_queries
    ADD CONSTRAINT error_code_queries_vehicle_token_id_fkey
        FOREIGN KEY (vehicle_token_id)
            REFERENCES user_devices (token_id)
                ON DELETE CASCADE;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
SET search_path = devices_api, public;
ALTER TABLE error_code_queries DROP CONSTRAINT error_code_queries_vehicle_token_id_fkey;
ALTER TABLE error_code_queries DROP COLUMN vehicle_token_id;
-- +goose StatementEnd
