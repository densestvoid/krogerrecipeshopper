-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS ingredients (
    product_id VARCHAR(13) NOT NULL,
    recipe_id UUID REFERENCES recipes(id),
    quantity INTEGER NOT NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS ingredients;
-- +goose StatementEnd
