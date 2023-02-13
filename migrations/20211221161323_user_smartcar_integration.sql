-- +goose Up
-- +goose StatementBegin
SET search_path = devices_api, public;

CREATE TYPE user_device_api_integration_status AS ENUM ('Pending', 'PendingFirstData', 'Active');

CREATE TABLE user_device_api_integrations (
    user_device_id CHAR(27) NOT NULL,
    integration_id CHAR(27) NOT NULL,
    status user_device_api_integration_status NOT NULL,
    access_token TEXT NOT NULL,
    access_expires_at timestamptz NOT NULL,
    refresh_token TEXT NOT NULL,
    refresh_expires_at timestamptz NOT NULL,

    PRIMARY KEY (user_device_id, integration_id),
    CONSTRAINT fkey_user_device_id FOREIGN KEY (user_device_id) REFERENCES user_devices(id),
    CONSTRAINT fkey_integration_id FOREIGN KEY (integration_id) REFERENCES integrations(id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SET search_path = devices_api, public;

DROP TABLE user_device_api_integrations;
DROP TYPE IF EXISTS user_device_api_integration_status;
-- +goose StatementEnd
