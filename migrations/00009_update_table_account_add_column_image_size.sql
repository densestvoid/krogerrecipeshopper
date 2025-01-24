-- +goose Up
-- +goose StatementBegin
CREATE TYPE image_size AS ENUM (
    'thumbnail',
    'small',
    'medium',
    'large',
    'xlarge'
);

ALTER TABLE accounts
    ADD COLUMN image_size image_size NOT NULL DEFAULT 'medium';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE accounts
    DROP COLUMN image_size;

DROP TYPE IF EXISTS image_size; 
-- +goose StatementEnd
