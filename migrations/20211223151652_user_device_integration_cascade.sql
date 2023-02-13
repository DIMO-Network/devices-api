-- +goose Up
-- +goose StatementBegin
SET search_path = devices_api, public;

ALTER TABLE user_device_api_integrations DROP CONSTRAINT fkey_user_device_id;
ALTER TABLE user_device_api_integrations ADD CONSTRAINT fkey_user_device_id FOREIGN KEY (user_device_id) REFERENCES user_devices(id) ON DELETE CASCADE;

ALTER TABLE user_device_api_integrations DROP CONSTRAINT fkey_integration_id;
ALTER TABLE user_device_api_integrations ADD CONSTRAINT fkey_integration_id FOREIGN KEY (integration_id) REFERENCES integrations(id) ON DELETE CASCADE;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SET search_path = devices_api, public;

ALTER TABLE user_device_api_integrations DROP CONSTRAINT fkey_user_device_id;
ALTER TABLE user_device_api_integrations ADD CONSTRAINT fkey_user_device_id FOREIGN KEY (user_device_id) REFERENCES user_devices(id);

ALTER TABLE user_device_api_integrations DROP CONSTRAINT fkey_integration_id;
ALTER TABLE user_device_api_integrations ADD CONSTRAINT fkey_integration_id FOREIGN KEY (integration_id) REFERENCES integrations(id);
-- +goose StatementEnd
