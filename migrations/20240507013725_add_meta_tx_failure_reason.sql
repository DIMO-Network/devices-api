-- +goose Up
-- +goose StatementBegin
ALTER TABLE meta_transaction_requests
    ADD COLUMN failure_reason TEXT;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE meta_transaction_requests
    DROP COLUMN failure_reason;
-- +goose StatementEnd
