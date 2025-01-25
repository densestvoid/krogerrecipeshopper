-- +goose Up
-- +goose StatementBegin
CREATE TYPE visibility AS ENUM (
    'private',
    'friends',
    'public'
);

ALTER TABLE recipes
    ADD COLUMN visibility visibility NOT NULL DEFAULT 'public';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE recipes
    DROP COLUMN visibility;

DROP TYPE visibility;
-- +goose StatementEnd
