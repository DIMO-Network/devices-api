-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
drop table user_device_data;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
