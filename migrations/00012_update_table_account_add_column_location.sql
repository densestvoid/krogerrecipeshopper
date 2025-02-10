-- +goose Up
-- +goose StatementBegin
ALTER TABLE accounts
    ADD COLUMN location_id VARCHAR(8) DEFAULT NULL
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE accounts
    DROP COLUMN location_id
-- +goose StatementEnd
