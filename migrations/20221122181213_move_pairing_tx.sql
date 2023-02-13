-- +goose Up
-- +goose StatementBegin
SET search_path = devices_api, public;

ALTER TABLE autopi_units
    ADD COLUMN pair_request_id char(27)
        CONSTRAINT autopi_units_pair_request_id_key UNIQUE
        CONSTRAINT autopi_units_pair_request_id_fkey REFERENCES meta_transaction_requests(id),
    ADD COLUMN unpair_request_id char(27)
        CONSTRAINT autopi_units_unpair_request_id_key UNIQUE
        CONSTRAINT autopi_units_unpair_request_id_fkey REFERENCES meta_transaction_requests(id);

UPDATE autopi_units AS au
    SET
        pair_request_id = udai.pair_meta_transaction_request_id,
        unpair_request_id = udai.unpair_meta_transaction_request_id
    FROM user_device_api_integrations AS udai 
    WHERE udai.autopi_unit_id = au.autopi_unit_id;

ALTER TABLE user_device_api_integrations
    DROP COLUMN pair_meta_transaction_request_id,
    DROP COLUMN unpair_meta_transaction_request_id;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SET search_path = devices_api, public;

ALTER TABLE user_device_api_integrations
    ADD COLUMN pair_meta_transaction_request_id char(27)
        CONSTRAINT user_device_api_integrations_pair_meta_transaction_request_id_key UNIQUE
        CONSTRAINT user_device_api_integrations_pair_meta_transaction_request_id_fkey REFERENCES meta_transaction_requests(id),
    ADD COLUMN unpair_meta_transaction_request_id char(27)
        CONSTRAINT user_device_api_integrations_unpair_transaction_request_id_key UNIQUE
        CONSTRAINT user_device_api_integrations_unpair_transaction_request_id_fkey REFERENCES meta_transaction_requests(id);

UPDATE user_device_api_integrations AS udai
    SET
        pair_meta_transaction_request_id = au.pair_request_id,
        unpair_meta_transaction_request_id = au.unpair_request_id
    FROM autopi_units AS au 
    WHERE udai.autopi_unit_id = au.autopi_unit_id;

ALTER TABLE autopi_units
    DROP COLUMN pair_request_id,
    DROP COLUMN unpair_request_id;
-- +goose StatementEnd
