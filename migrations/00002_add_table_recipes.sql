-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS recipes (
    id UUID PRIMARY KEY NOT NULL DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL,
    name VARCHAR(256) NOT NULL,
    description VARCHAR(1024)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS recipes;
-- +goose StatementEnd
