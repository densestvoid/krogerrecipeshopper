package data

import (
	"context"

	"github.com/google/uuid"
)

type Ingredient struct {
	ProductID string
	RecipeID  uuid.UUID
	Quantity  int // represents a percentage of the total product
	Staple    bool
}

func (m *Repository) GetIngredient(ctx context.Context, recipeID uuid.UUID, productID string) (*Ingredient, error) {
	row := m.db.QueryRowContext(ctx, `SELECT product_id, recipe_id, quantity, staple FROM ingredients WHERE product_id = $1 AND recipe_id = $2`, productID, recipeID)
	if err := row.Err(); err != nil {
		return nil, err
	}
	var ingredient Ingredient
	return &ingredient, row.Scan(&ingredient.ProductID, &ingredient.RecipeID, &ingredient.Quantity, &ingredient.Staple)
}

func (m *Repository) ListIngredients(ctx context.Context, recipeID uuid.UUID) ([]Ingredient, error) {
	rows, err := m.db.QueryContext(ctx, `SELECT product_id, recipe_id, quantity, staple FROM ingredients WHERE recipe_id = $1 ORDER BY staple`, recipeID)
	if err != nil {
		return nil, err
	}
	var ingredients []Ingredient
	for rows.Next() {
		var ingredient Ingredient
		if err := rows.Scan(&ingredient.ProductID, &ingredient.RecipeID, &ingredient.Quantity, &ingredient.Staple); err != nil {
			return nil, err
		}
		ingredients = append(ingredients, ingredient)
	}
	return ingredients, nil
}

func (m *Repository) CreateIngredient(ctx context.Context, productID string, recipeID uuid.UUID, quantity int, staple bool) error {
	row := m.db.QueryRowContext(ctx, `INSERT INTO ingredients(product_id, recipe_id , quantity, staple) VALUES ($1, $2, $3, $4)`, productID, recipeID, quantity, staple)
	return row.Err()
}

func (m *Repository) UpdateIngredient(ctx context.Context, productID string, recipeID uuid.UUID, quantity int, staple bool) error {
	_, err := m.db.ExecContext(ctx, `UPDATE ingredients SET quantity=$1, staple=$2 WHERE product_id=$3 and recipe_id=$4`, quantity, staple, productID, recipeID)
	return err
}

func (m *Repository) DeleteIngredient(ctx context.Context, productID string, recipeID uuid.UUID) error {
	_, err := m.db.ExecContext(ctx, `DELETE FROM ingredients WHERE product_id=$1 and recipe_id=$2`, productID, recipeID)
	return err
}
