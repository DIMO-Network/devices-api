-- +goose Up
-- +goose StatementBegin
SET search_path = devices_api, public;
ALTER TABLE user_devices ADD COLUMN ipfs_image_cid varchar;
-- +goose StatementEnd


-- +goose Down
-- +goose StatementBegin
SET search_path = devices_api, public;
ALTER TABLE user_devices DROP COLUMN ipfs_image_cid;
-- +goose StatementEnd
