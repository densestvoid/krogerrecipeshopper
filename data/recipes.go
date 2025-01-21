package data

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

type Recipe struct {
	ID          uuid.UUID
	UserID      uuid.UUID
	Name        string
	Description string
}

func (m *Repository) GetRecipe(ctx context.Context, id uuid.UUID) (*Recipe, error) {
	row := m.db.QueryRowContext(ctx, `SELECT id, user_id, name, description FROM recipes WHERE id = $1`, id)
	if err := row.Err(); err != nil {
		return nil, err
	}
	var recipe Recipe
	return &recipe, row.Scan(&recipe.ID, &recipe.UserID, &recipe.Name, &recipe.Description)
}

type ListRecipesFilter interface {
	listRecipoesFilter() string
}

type ListRecipesFilterByUserID struct {
	UserID uuid.UUID
}

func (f ListRecipesFilterByUserID) listRecipoesFilter() string {
	return fmt.Sprintf(`user_id = '%v'`, f.UserID)
}

type ListRecipesFilterByName struct {
	Name string
}

func (f ListRecipesFilterByName) listRecipoesFilter() string {
	return fmt.Sprintf(`name ILIKE '%%%s%%'`, f.Name)
}

func (m *Repository) ListRecipes(ctx context.Context, filters ...ListRecipesFilter) ([]Recipe, error) {
	query := `SELECT id, user_id, name, description FROM recipes`
	if len(filters) > 0 {
		query += " WHERE "
		filterStrings := []string{}
		for _, filter := range filters {
			filterStrings = append(filterStrings, filter.listRecipoesFilter())
		}
		query += strings.Join(filterStrings, " AND ")
	}
	rows, err := m.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	var recipes []Recipe
	for rows.Next() {
		var recipe Recipe
		if err := rows.Scan(&recipe.ID, &recipe.UserID, &recipe.Name, &recipe.Description); err != nil {
			return nil, err
		}
		recipes = append(recipes, recipe)
	}
	return recipes, nil
}

func (m *Repository) CreateRecipe(ctx context.Context, krogerUserID uuid.UUID, name, description string) (uuid.UUID, error) {
	row := m.db.QueryRowContext(ctx, `INSERT INTO recipes(user_id, name, description) VALUES ($1, $2, $3) RETURNING id`, krogerUserID, name, description)
	if err := row.Err(); err != nil {
		return uuid.Nil, err
	}
	var id uuid.UUID
	return id, row.Scan(&id)
}

func (m *Repository) UpdateRecipe(ctx context.Context, recipe Recipe) error {
	_, err := m.db.ExecContext(ctx, `UPDATE recipes SET name=$1, description=$2 WHERE id=$3`, recipe.Name, recipe.Description, recipe.ID)
	return err
}

func (m *Repository) DeleteRecipe(ctx context.Context, id uuid.UUID) error {
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, `DELETE FROM ingredients where recipe_id = $1`, id); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM recipes WHERE id = $1`, id); err != nil {
		return err
	}
	return tx.Commit()
}
