-- +goose Up
-- +goose StatementBegin
DROP TABLE IF EXISTS user_device_data;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
CREATE TABLE user_device_data(
    user_device_id char(27) NOT NULL,
    integration_id char(27) NOT NULL,
    signals jsonb,
    error_data jsonb,
    last_odometer_event_at timestamptz,
    real_last_odometer_event_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT current_timestamp,
    updated_at timestamptz NOT NULL DEFAULT current_timestamp,
    CONSTRAINT user_device_data_pkey PRIMARY KEY (user_device_id, integration_id)
);
-- +goose StatementEnd
