-- +goose Up
-- +goose StatementBegin
-- create the account table
CREATE TABLE IF NOT EXISTS accounts (
    id UUID PRIMARY KEY NOT NULL DEFAULT uuid_generate_v4(),
    kroger_profile_id UUID NOT NULL
);

-- create accounts for user_ids in recipes, favorites, or cart products
INSERT INTO accounts (kroger_profile_id)
SELECT DISTINCT user_id
FROM (
    SELECT user_id FROM recipes
    UNION ALL
    SELECT user_id FROM favorites
    UNION ALL
    SELECT user_id FROM cart_products
);

-- add account_ids to recipes, favorites, and cart products
ALTER TABLE recipes
    ADD COLUMN account_id UUID;

ALTER TABLE favorites
    ADD COLUMN account_id UUID;

ALTER TABLE cart_products
    ADD COLUMN account_id UUID,
    ADD CONSTRAINT account_product UNIQUE (account_id, product_id);

-- update recipes, favorites, and cart products account_ids
UPDATE recipes
SET account_id = accounts.id
FROM accounts
WHERE recipes.user_id = accounts.kroger_profile_id;

UPDATE favorites
SET account_id = accounts.id
FROM accounts
WHERE favorites.user_id = accounts.kroger_profile_id;

UPDATE cart_products
SET account_id = accounts.id
FROM accounts
WHERE cart_products.user_id = accounts.kroger_profile_id;

-- enforce account id reference on account_id columns
ALTER TABLE recipes
    ALTER COLUMN account_id SET NOT NULL,
    ADD FOREIGN KEY (account_id) REFERENCES accounts(id);

ALTER TABLE favorites
    ALTER COLUMN account_id SET NOT NULL,
    ADD FOREIGN KEY (account_id) REFERENCES accounts(id);

ALTER TABLE cart_products
    ALTER COLUMN account_id SET NOT NULL,
    ADD FOREIGN KEY (account_id) REFERENCES accounts(id);

-- drop user_id from recipes, favorites, and cart products
ALTER TABLE recipes
    DROP COLUMN user_id;

ALTER TABLE favorites
    DROP COLUMN user_id;

ALTER TABLE cart_products
    DROP COLUMN user_id;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- add user_id to recipes, favorites, and cart products
ALTER TABLE recipes
    ADD COLUMN user_id UUID;

ALTER TABLE favorites
    ADD COLUMN user_id UUID;

ALTER TABLE cart_products
    ADD COLUMN user_id UUID,
    ADD CONSTRAINT user_product UNIQUE (user_id, product_id);

-- set user_id on recipes, favorites, and cart products
UPDATE recipes
SET user_id = accounts.kroger_profile_id
FROM accounts
WHERE recipes.account_id = accounts.id;

UPDATE favorites
SET user_id = accounts.kroger_profile_id
FROM accounts
WHERE favorites.account_id = accounts.id;

UPDATE cart_products
SET user_id = accounts.kroger_profile_id
FROM accounts
WHERE cart_products.account_id = accounts.id;


-- enforce not null on user_id columns
ALTER TABLE recipes
    ALTER COLUMN user_id SET NOT NULL;

ALTER TABLE favorites
    ALTER COLUMN user_id SET NOT NULL;

ALTER TABLE cart_products
    ALTER COLUMN user_id SET NOT NULL;
    
-- drop account_id from recipes, favorites, and cart products
ALTER TABLE recipes
    DROP COLUMN account_id;

ALTER TABLE favorites
    DROP COLUMN account_id;

ALTER TABLE cart_products
    DROP COLUMN account_id;

-- drop the accounts table
DROP TABLE accounts;
-- +goose StatementEnd
