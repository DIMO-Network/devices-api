-- +goose Up
-- +goose StatementBegin
SET search_path = devices_api, public;

CREATE TYPE device_command_request_status AS ENUM ('Pending', 'Complete', 'Failed');

CREATE TABLE device_command_requests (
    id char(27) NOT NULL,
    user_device_id char(27) NOT NULL,
    integration_id char(27) NOT NULL,
    command TEXT NOT NULL,
    status device_command_request_status NOT NULL,
    created_at timestamptz NOT NULL DEFAULT current_timestamp,
    updated_at timestamptz NOT NULL DEFAULT current_timestamp,
    CONSTRAINT device_command_requests_id_pkey PRIMARY KEY(id),
    CONSTRAINT device_command_requests_user_device_id_fkey FOREIGN KEY (user_device_id) REFERENCES user_devices(id) ON DELETE CASCADE,
    CONSTRAINT device_command_requests_integration_id_fkey FOREIGN KEY (integration_id) REFERENCES integrations(id) ON DELETE CASCADE
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SET search_path = devices_api, public;

DROP TABLE device_command_requests;
DROP TYPE device_command_request_status;
-- +goose StatementEnd
