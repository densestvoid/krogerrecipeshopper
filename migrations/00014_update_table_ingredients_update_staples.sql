-- +goose Up
-- +goose StatementBegin
UPDATE ingredients
SET quantity = 1
WHERE staple = true;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- NO WAY TO REVERSE THIS OPERATION
-- +goose StatementEnd
