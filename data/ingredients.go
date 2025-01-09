package data

import (
	"context"

	"github.com/google/uuid"
)

type Ingredient struct {
	ProductID string
	RecipeID  uuid.UUID
	Quantity  int // represents a percentage of the total product
}

func (m *Repository) ListIngredients(ctx context.Context, recipeID uuid.UUID) ([]Ingredient, error) {
	rows, err := m.db.QueryContext(ctx, `SELECT product_id, recipe_id, quantity FROM ingredients WHERE recipe_id = $1`, recipeID)
	if err != nil {
		return nil, err
	}
	var ingredients []Ingredient
	for rows.Next() {
		var ingredient Ingredient
		if err := rows.Scan(&ingredient.ProductID, &ingredient.RecipeID, &ingredient.Quantity); err != nil {
			return nil, err
		}
		ingredients = append(ingredients, ingredient)
	}
	return ingredients, nil
}

func (m *Repository) CreateIngredient(ctx context.Context, productID string, recipeID uuid.UUID, quantity int) (uuid.UUID, error) {
	row := m.db.QueryRowContext(ctx, `INSERT INTO ingredients(productID, recipeID, quantity) VALUES ($1, $2, $3) RETURNING id`, productID, recipeID, quantity)
	if err := row.Err(); err != nil {
		return uuid.Nil, err
	}
	var id uuid.UUID
	return id, row.Scan(&id)
}

func (m *Repository) UpdateIngredient(ctx context.Context, productID string, recipeID uuid.UUID, quantity int) error {
	_, err := m.db.ExecContext(ctx, `UPDATE ingredients SET quantity=$1 WHERE productID=$2 and recipeID=$3`, quantity, productID, recipeID)
	return err
}

func (m *Repository) DeleteIngredient(ctx context.Context, productID string, recipeID uuid.UUID) error {
	_, err := m.db.ExecContext(ctx, `DELETE FROM ingredients WHERE productID=$1 and recipeID=$2`, productID, recipeID)
	return err
}
