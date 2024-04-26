-- +goose Up
-- +goose StatementBegin
ALTER TYPE meta_transaction_request_status ADD VALUE 'Failed';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- No great way to subtract from an enum.
-- +goose StatementEnd
