-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
-- +goose StatementEnd
SET search_path = devices_api, public;
UPDATE user_devices 
    SET 
        user_id = condensed.user_ids[1]
FROM 
    (
        SELECT
            ud.owner_address,
            ARRAY_AGG(distinct ud.user_id) user_ids
        FROM
            user_devices ud
        GROUP BY ud.owner_address
    ) condensed 
WHERE condensed.owner_address = user_devices.owner_address;

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
