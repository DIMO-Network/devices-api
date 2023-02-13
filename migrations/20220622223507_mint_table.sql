-- +goose Up
-- +goose StatementBegin
SET search_path = devices_api, public;

CREATE TYPE txstate AS ENUM ('Unstarted', 'Submitted', 'Mined', 'Confirmed');

CREATE TABLE mint_requests (
    id char(27) PRIMARY KEY,
    user_device_id char(27) NOT NULL,
    tx_state txstate NOT NULL DEFAULT 'Unstarted',
    token_id numeric(78, 0),
    created_at timestamptz NOT NULL DEFAULT current_timestamp,
    updated_at timestamptz NOT NULL DEFAULT current_timestamp
);

ALTER TABLE mint_requests ADD CONSTRAINT mint_requests_user_device_id_fkey FOREIGN KEY (user_device_id) REFERENCES user_devices(id);

ALTER TABLE user_devices ADD COLUMN token_id numeric(78, 0);
ALTER TABLE user_devices ADD CONSTRAINT user_devices_token_id_key UNIQUE (token_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SET search_path = devices_api, public;

DROP TABLE mint_requests;
DROP TYPE txstate;
-- +goose StatementEnd
