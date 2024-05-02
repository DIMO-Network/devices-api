-- +goose Up
-- +goose StatementBegin
ALTER TABLE autopi_jobs
    DROP CONSTRAINT autopi_jobs_aftermarket_devices,
    ADD CONSTRAINT autopi_jobs_autopi_unit_id_fkey FOREIGN KEY (autopi_unit_id) REFERENCES aftermarket_devices(serial) ON DELETE CASCADE;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE autopi_jobs
    DROP CONSTRAINT autopi_jobs_autopi_unit_id_fkey,
    ADD CONSTRAINT autopi_jobs_aftermarket_devices FOREIGN KEY (autopi_unit_id) REFERENCES aftermarket_devices(serial);
-- +goose StatementEnd
