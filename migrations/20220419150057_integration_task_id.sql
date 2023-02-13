-- +goose Up
-- +goose StatementBegin
ALTER TABLE user_device_api_integrations ADD COLUMN task_id char(27);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE user_device_api_integrations DROP COLUMN task_id;
-- +goose StatementEnd
