-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';

ALTER TABLE vehicle_nfts
  RENAME COLUMN mint_request_id TO transaction_request_id;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';

ALTER TABLE vehicle_nfts
  RENAME COLUMN transaction_request_id TO mint_request_id;

-- +goose StatementEnd
