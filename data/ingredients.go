package data

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

type Ingredient struct {
	ProductID string    `db:"product_id"`
	ListID    uuid.UUID `db:"list_id"`
	Quantity  int       `db:"quantity"` // represents a percentage of the total product
	Staple    bool      `db:"staple"`
}

func (i Ingredient) QuantityDecimalString() string {
	if i.Quantity == 0 {
		return ""
	}
	return fmt.Sprintf("%.2f", float64(i.Quantity)/100)
}

func (r *Repository) GetIngredient(ctx context.Context, listID uuid.UUID, productID string) (Ingredient, error) {
	var ingredient Ingredient
	return ingredient, r.db.GetContext(ctx, &ingredient, `SELECT product_id, list_id, quantity, staple FROM ingredients WHERE product_id = $1 AND list_id = $2`, productID, listID)
}

func (m *Repository) ListIngredients(ctx context.Context, listID uuid.UUID) ([]Ingredient, error) {
	ingredients := []Ingredient{}
	return ingredients, m.db.SelectContext(ctx, &ingredients, `SELECT product_id, list_id, quantity, staple FROM ingredients WHERE list_id = $1 ORDER BY staple`, listID)
}

func (m *Repository) CreateIngredient(ctx context.Context, productID string, listID uuid.UUID, quantity int, staple bool) error {
	row := m.db.QueryRowContext(ctx, `INSERT INTO ingredients(product_id, list_id , quantity, staple) VALUES ($1, $2, $3, $4)`, productID, listID, quantity, staple)
	return row.Err()
}

func (m *Repository) UpdateIngredient(ctx context.Context, productID string, listID uuid.UUID, quantity int, staple bool) error {
	_, err := m.db.ExecContext(ctx, `UPDATE ingredients SET quantity=$1, staple=$2 WHERE product_id=$3 and list_id=$4`, quantity, staple, productID, listID)
	return err
}

func (m *Repository) DeleteIngredient(ctx context.Context, productID string, listID uuid.UUID) error {
	_, err := m.db.ExecContext(ctx, `DELETE FROM ingredients WHERE product_id=$1 and list_id=$2`, productID, listID)
	return err
}
