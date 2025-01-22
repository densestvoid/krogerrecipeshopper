package data

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
)

type CartProduct struct {
	UserID    uuid.UUID
	ProductID string
	Quantity  int
	Staple    bool
}

func (r *Repository) GetCartProduct(ctx context.Context, userID uuid.UUID, productID string) (CartProduct, error) {
	row := r.db.QueryRowContext(ctx, `SELECT user_id, product_id, quantity, staple FROM cart_products WHERE user_id = $1 AND product_id = $2`, userID, productID)
	if err := row.Err(); err != nil {
		return CartProduct{}, err
	}
	var cartProduct CartProduct
	return cartProduct, row.Scan(&cartProduct.UserID, &cartProduct.ProductID, &cartProduct.Quantity, &cartProduct.Staple)
}

type ListCartProductsFilter interface {
	listCartProductsFilter() string
}

type ListCartProductsNonStaples struct{}

func (r *ListCartProductsNonStaples) listCartProductsFilter() string {
	return "staple = false"
}

func (r *Repository) ListCartProducts(ctx context.Context, userID uuid.UUID, filters ...ListCartProductsFilter) ([]*CartProduct, error) {
	query := `SELECT user_id, product_id, quantity, staple FROM cart_products WHERE user_id = $1`
	for _, filter := range filters {
		query += fmt.Sprintf(" AND %s", filter.listCartProductsFilter())
	}
	query += " ORDER BY staple"
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cartProducts []*CartProduct
	for rows.Next() {
		var cartProduct CartProduct
		err := rows.Scan(&cartProduct.UserID, &cartProduct.ProductID, &cartProduct.Quantity, &cartProduct.Staple)
		if err != nil {
			return nil, err
		}
		cartProducts = append(cartProducts, &cartProduct)
	}
	return cartProducts, nil
}

func (r *Repository) AddCartProduct(ctx context.Context, userID uuid.UUID, productID string, quantity int, staple bool) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO cart_products (user_id, product_id, quantity, staple)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (user_id, product_id)
		DO UPDATE SET quantity = cart_products.quantity + EXCLUDED.quantity;
	`, userID, productID, quantity, staple)
	return err
}

func (r *Repository) SetCartProduct(ctx context.Context, userID uuid.UUID, productID string, quantity *int, staple *bool) error {
	dbConn, err := r.db.Conn(ctx)
	if err != nil {
		return err
	}
	return dbConn.Raw(func(driverConn any) error {
		conn := driverConn.(*stdlib.Conn).Conn() // conn is a *pgx.Conn
		args := pgx.NamedArgs{
			"userID":    userID,
			"productID": productID,
		}
		sets := []string{}
		if quantity != nil {
			args["quantity"] = *quantity
			sets = append(sets, "quantity = @quantity")
		}
		if staple != nil {
			args["staple"] = *staple
			sets = append(sets, "staple = @staple")
		}
		if len(sets) == 0 {
			return nil
		}
		query := fmt.Sprintf(`UPDATE cart_products SET %s WHERE user_id = @userID AND product_id = @productID;`, strings.Join(sets, ", "))
		_, err := conn.Exec(ctx, query, args)
		return err
	})
}

func (r *Repository) RemoveCartProduct(ctx context.Context, userID uuid.UUID, productID string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM cart_products WHERE user_id = $1 and product_id = $2;`, userID, productID)
	return err
}

func (r *Repository) ClearCartProducts(ctx context.Context, userID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM cart_products WHERE user_id = $1;`, userID)
	return err
}
