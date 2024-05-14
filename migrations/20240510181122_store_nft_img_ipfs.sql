-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
-- +goose StatementEnd

SET search_path = devices_api, public;
ALTER TABLE user_devices ADD COLUMN ipfs_image_cid VARCHAR;

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';

SET search_path = devices_api, public;
ALTER TABLE user_devices DROP COLUMN ipfs_image_cid;
-- +goose StatementEnd
