-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS tags (
    id UUID PRIMARY KEY NOT NULL DEFAULT uuid_generate_v4(),
    name VARCHAR(32) NOT NULL
);

CREATE UNIQUE INDEX tags_name_lower_unique ON tags (lower(name));

CREATE TABLE IF NOT EXISTS recipe_tags (
    recipe_list_id UUID NOT NULL REFERENCES recipes (list_id),
    tag_id UUID NOT NULL REFERENCES tags (id),
    PRIMARY KEY (recipe_list_id, tag_id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS recipe_tags;
DROP TABLE IF EXISTS tags;
-- +goose StatementEnd
