package data

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

type CartProduct struct {
	AccountID uuid.UUID
	ProductID string
	Quantity  int
	Staple    bool
}

func (r *Repository) GetCartProduct(ctx context.Context, accountID uuid.UUID, productID string) (CartProduct, error) {
	row := r.db.QueryRowContext(ctx, `SELECT account_id, product_id, quantity, staple FROM cart_products WHERE account_id = $1 AND product_id = $2`, accountID, productID)
	if err := row.Err(); err != nil {
		return CartProduct{}, err
	}
	var cartProduct CartProduct
	return cartProduct, row.Scan(&cartProduct.AccountID, &cartProduct.ProductID, &cartProduct.Quantity, &cartProduct.Staple)
}

type ListCartProductsFilter interface {
	listCartProductsFilter() string
}

type ListCartProductsIncludeStaples struct {
	Include bool
}

func (r *ListCartProductsIncludeStaples) listCartProductsFilter() string {
	return fmt.Sprintf("staple = %t", r.Include)
}

func (r *Repository) ListCartProducts(ctx context.Context, accountID uuid.UUID, filters ...ListCartProductsFilter) ([]*CartProduct, error) {
	query := `SELECT account_id, product_id, quantity, staple FROM cart_products WHERE account_id = $1`
	for _, filter := range filters {
		query += fmt.Sprintf(" AND %s", filter.listCartProductsFilter())
	}
	query += " ORDER BY staple"
	rows, err := r.db.QueryContext(ctx, query, accountID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cartProducts []*CartProduct
	for rows.Next() {
		var cartProduct CartProduct
		err := rows.Scan(&cartProduct.AccountID, &cartProduct.ProductID, &cartProduct.Quantity, &cartProduct.Staple)
		if err != nil {
			return nil, err
		}
		cartProducts = append(cartProducts, &cartProduct)
	}
	return cartProducts, nil
}

func (r *Repository) AddCartProduct(ctx context.Context, accountID uuid.UUID, productID string, quantity int, staple bool) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO cart_products (account_id, product_id, quantity, staple)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (account_id, product_id)
		DO UPDATE SET quantity = cart_products.quantity + CASE
			WHEN EXCLUDED.staple THEN 0
			WHEN not EXCLUDED.staple THEN EXCLUDED.quantity
		END;
	`, accountID, productID, quantity, staple)
	return err
}

func (r *Repository) SetCartProduct(ctx context.Context, accountID uuid.UUID, productID string, quantity *int, staple *bool) error {
	namedArgs := map[string]any{
		"accountID": accountID,
		"productID": productID,
	}

	sets := []string{}
	if quantity != nil {
		namedArgs["quantity"] = *quantity
		sets = append(sets, "quantity = :quantity")
	}
	if staple != nil {
		namedArgs["staple"] = *staple
		sets = append(sets, "staple = :staple")
	}
	if len(sets) == 0 {
		return nil
	}

	query := fmt.Sprintf(`UPDATE cart_products SET %s WHERE account_id = :accountID AND product_id = :productID;`, strings.Join(sets, ", "))
	namedStmt, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		return err
	}
	defer namedStmt.Close()

	_, err = namedStmt.ExecContext(ctx, namedArgs)
	return err
}

func (r *Repository) RemoveCartProduct(ctx context.Context, accountID uuid.UUID, productID string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM cart_products WHERE account_id = $1 and product_id = $2;`, accountID, productID)
	return err
}

func (r *Repository) ClearCartProducts(ctx context.Context, accountID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM cart_products WHERE account_id = $1;`, accountID)
	return err
}
