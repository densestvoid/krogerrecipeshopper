-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS cart_products (
    user_id UUID NOT NULL,
    product_id VARCHAR(13) NOT NULL,
    quantity INTEGER NOT NULL,
    CONSTRAINT user_product UNIQUE (user_id, product_id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS cart_products;
-- +goose StatementEnd
