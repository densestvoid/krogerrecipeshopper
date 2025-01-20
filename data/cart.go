package data

import (
	"context"

	"github.com/google/uuid"
)

type CartProduct struct {
	UserID    uuid.UUID
	ProductID string
	Quantity  int
}

func (r *Repository) GetCartProduct(ctx context.Context, userID uuid.UUID, productID string) (CartProduct, error) {
	row := r.db.QueryRowContext(ctx, `SELECT user_id, product_id, quantity FROM cart_products WHERE user_id = $1 AND product_id = $2`, userID, productID)
	if err := row.Err(); err != nil {
		return CartProduct{}, err
	}
	var cartProduct CartProduct
	return cartProduct, row.Scan(&cartProduct.UserID, &cartProduct.ProductID, &cartProduct.Quantity)
}

func (r *Repository) ListCartProducts(ctx context.Context, userID uuid.UUID) ([]*CartProduct, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT user_id, product_id, quantity FROM cart_products WHERE user_id = $1`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cartProducts []*CartProduct
	for rows.Next() {
		var cartProduct CartProduct
		err := rows.Scan(&cartProduct.UserID, &cartProduct.ProductID, &cartProduct.Quantity)
		if err != nil {
			return nil, err
		}
		cartProducts = append(cartProducts, &cartProduct)
	}
	return cartProducts, nil
}

func (r *Repository) AddCartProduct(ctx context.Context, userID uuid.UUID, productID string, quantity int) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO cart_products (user_id, product_id, quantity)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id, product_id)
		DO UPDATE SET quantity = cart_products.quantity + EXCLUDED.quantity;
	`, userID, productID, quantity)
	return err
}

func (r *Repository) SetCartProduct(ctx context.Context, userID uuid.UUID, productID string, quantity int) error {
	_, err := r.db.ExecContext(ctx, `UPDATE cart_products SET quantity = $3 WHERE user_id = $1 and product_id = $2;`, userID, productID, quantity)
	return err
}

func (r *Repository) RemoveCartProduct(ctx context.Context, userID uuid.UUID, productID string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM cart_products WHERE user_id = $1 and product_id = $2;`, userID, productID)
	return err
}

func (r *Repository) ClearCartProducts(ctx context.Context, userID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM cart_products WHERE user_id = $1;`, userID)
	return err
}
