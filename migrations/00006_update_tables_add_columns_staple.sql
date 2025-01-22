-- +goose Up
-- +goose StatementBegin
ALTER TABLE ingredients
    ADD COLUMN staple BOOLEAN NOT NULL DEFAULT false;
ALTER TABLE cart_products
    ADD COLUMN staple BOOLEAN NOT NULL DEFAULT false;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE ingredients
    DROP COLUMN staple;
ALTER TABLE cart_products
    DROP COLUMN staple;
-- +goose StatementEnd
