-- +goose Up
-- +goose StatementBegin

-- create the lists table
CREATE TABLE IF NOT EXISTS lists (
    id UUID PRIMARY KEY NOT NULL DEFAULT uuid_generate_v4(),
    account_id UUID NOT NULL REFERENCES accounts (id),
    name VARCHAR(256) NOT NULL,
    description VARCHAR(1024)
);

INSERT INTO lists (id, account_id, name, description)
SELECT id, account_id, name, description FROM recipes;

-- update the recipes table
ALTER TABLE recipes
    ADD COLUMN list_id UUID UNIQUE REFERENCES lists (id);

UPDATE recipes SET list_id = id;

-- update the ingredients table
ALTER TABLE ingredients
    ADD COLUMN list_id UUID REFERENCES lists (id);

UPDATE ingredients SET list_id = recipe_id;

ALTER TABLE ingredients
    DROP COLUMN recipe_id,
    ALTER COLUMN list_id SET NOT NULL;

-- update the favorites table
ALTER TABLE favorites
    ADD COLUMN list_id UUID REFERENCES recipes (list_id);

UPDATE favorites SET list_id = recipe_id;

ALTER TABLE favorites
    DROP COLUMN recipe_id,
    ALTER COLUMN list_id SET NOT NULL;

-- drop the columns from the recipe table
ALTER TABLE recipes
    DROP COLUMN id,
    DROP COLUMN name,
    DROP COLUMN description,
    DROP COLUMN account_id,
    ALTER COLUMN list_id SET NOT NULL,
    ADD CONSTRAINT recipes_pkey PRIMARY KEY (list_id);

-- create recipe table view
CREATE OR REPLACE VIEW recipe_list_view AS
(
    SELECT
        list_id,
        lists.account_id AS account_id,
        lists.name,
        lists.description, 
        instruction_type,
        instructions,
        visibility
    FROM recipes
        INNER JOIN lists ON lists.id = recipes.list_id
);

-- +goose StatementEnd


-- +goose Down
-- +goose StatementBegin

-- drop recipe table view
DROP VIEW recipe_list_view;

-- add dropped columns to the recipes table
ALTER TABLE recipes
    ADD COLUMN id UUID PRIMARY KEY NOT NULL DEFAULT uuid_generate_v4(),
    ADD COLUMN account_id UUID REFERENCES accounts (id),
    ADD COLUMN name VARCHAR(256),
    ADD COLUMN description VARCHAR(1024);

-- copy list details to recipe
UPDATE recipes
SET id = list_id, account_id = l.account_id, name = l.name, description = l.description
FROM lists l
WHERE recipes.list_id = l.id;

-- downgrade favorites table
ALTER TABLE favorites
    ADD COLUMN recipe_id REFERENCES recipes (id);

UPDATE favorites SET recipe_id = list_id;
    
ALTER TABLE favorites
    DROP COLUMN list_id,
    ALTER COLUMN recipe_id SET NOT NULL;

-- downgrade ingredients table
ALTER TABLE ingredients
    ADD COLUMN recipe_id REFERENCES recipes (id);

UPDATE ingredients SET recipe_id = list_id;
    
ALTER TABLE ingredients
    DROP COLUMN list_id,
    ALTER COLUMN recipe_id SET NOT NULL;

-- downgrade the recipes table
ALTER TABLE recipes
    DROP COLUMN list_id,
    ALTER COLUMN account_id SET NOT NULL,
    ALTER COLUMN name SET NOT NULL,
    ADD CONSTRAINT recipes_pkey PRIMARY KEY (id);

-- drop the list table
DROP TABLE lists;

-- +goose StatementEnd