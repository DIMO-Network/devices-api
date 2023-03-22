-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';

ALTER TABLE error_code_queries DROP CONSTRAINT error_code_queries_user_device_id_fkey;
ALTER TABLE error_code_queries ADD CONSTRAINT error_code_queries_user_device_id_fkey FOREIGN KEY (user_device_id) REFERENCES user_devices(id) ON DELETE CASCADE;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';

ALTER TABLE error_code_queries DROP CONSTRAINT error_code_queries_user_device_id_fkey;
ALTER TABLE error_code_queries ADD CONSTRAINT error_code_queries_user_device_id_fkey FOREIGN KEY (user_device_id) REFERENCES user_devices(id);
-- +goose StatementEnd
