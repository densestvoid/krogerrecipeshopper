-- +goose Up
-- +goose StatementBegin
CREATE TYPE instruction_type AS ENUM (
    'none',
    'text',
    'link'
);

ALTER TABLE recipes
    ADD COLUMN instruction_type instruction_type NOT NULL DEFAULT 'text',
    ADD COLUMN instructions TEXT NOT NULL  DEFAULT '';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE recipes
    DROP COLUMN instruction_type,
    DROP COLUMN instructions;

DROP TYPE instruction_type;
-- +goose StatementEnd
