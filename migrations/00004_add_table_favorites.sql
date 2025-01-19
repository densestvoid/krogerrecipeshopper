-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS favorites (
    user_id UUID NOT NULL,
    recipe_id UUID REFERENCES recipes(id)
)
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS favorites;
-- +goose StatementEnd
