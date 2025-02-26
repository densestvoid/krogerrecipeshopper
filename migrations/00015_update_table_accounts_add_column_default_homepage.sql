-- +goose Up
-- +goose StatementBegin
CREATE TYPE homepage_option AS ENUM (
    'welcome',
    'recipes',
    'favorites',
    'explore'
);

ALTER TABLE accounts
    ADD COLUMN homepage homepage_option NOT NULL DEFAULT 'welcome';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE accounts
    DROP COLUMN homepage;

DROP TYPE homepage_option;
-- +goose StatementEnd
