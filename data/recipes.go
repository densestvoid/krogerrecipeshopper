package data

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

const (
	VisibilityPrivate = "private"
	VisibilityFriends = "friends"
	VisibilityPublic  = "public"
)

type Recipe struct {
	ID          uuid.UUID
	AccountID   uuid.UUID
	Name        string
	Description string
	Visibility  string
}

func (m *Repository) GetRecipe(ctx context.Context, id uuid.UUID) (*Recipe, error) {
	row := m.db.QueryRowContext(ctx, `SELECT id, account_id, name, description, visibility FROM recipes WHERE id = $1`, id)
	if err := row.Err(); err != nil {
		return nil, err
	}
	var recipe Recipe
	return &recipe, row.Scan(&recipe.ID, &recipe.AccountID, &recipe.Name, &recipe.Description, &recipe.Visibility)
}

type ListRecipesFilter interface {
	listRecipoesFilter() string
}

type ListRecipesFilterByAccountID struct {
	AccountID uuid.UUID
}

func (f ListRecipesFilterByAccountID) listRecipoesFilter() string {
	return fmt.Sprintf(`account_id = '%v'`, f.AccountID)
}

type ListRecipesFilterByName struct {
	Name string
}

func (f ListRecipesFilterByName) listRecipoesFilter() string {
	return fmt.Sprintf(`name ILIKE '%%%s%%'`, f.Name)
}

type ListRecipesOrderBy struct {
	Field     string
	Direction string
}

func (m *Repository) ListRecipes(ctx context.Context, accountID uuid.UUID, filters []ListRecipesFilter, orderBys []ListRecipesOrderBy) ([]Recipe, error) {
	query := `SELECT id, account_id, name, description, visibility FROM recipes WHERE (account_id = $1 OR visibility = 'public')`
	if len(filters) > 0 {
		query += " AND "
		filterStrings := []string{}
		for _, filter := range filters {
			filterStrings = append(filterStrings, filter.listRecipoesFilter())
		}
		query += strings.Join(filterStrings, " AND ")
	}
	if len(orderBys) > 0 {
		query += " ORDER BY "
		orderStrings := []string{}
		for _, orderBy := range orderBys {
			orderStrings = append(orderStrings, fmt.Sprintf("%s %s", orderBy.Field, orderBy.Direction))
		}
		query += strings.Join(orderStrings, ", ")
	}

	rows, err := m.db.QueryContext(ctx, query, accountID)
	if err != nil {
		return nil, err
	}
	var recipes []Recipe
	for rows.Next() {
		var recipe Recipe
		if err := rows.Scan(&recipe.ID, &recipe.AccountID, &recipe.Name, &recipe.Description, &recipe.Visibility); err != nil {
			return nil, err
		}
		recipes = append(recipes, recipe)
	}
	return recipes, nil
}

func (m *Repository) CreateRecipe(ctx context.Context, accountID uuid.UUID, name, description, visibility string) (uuid.UUID, error) {
	row := m.db.QueryRowContext(ctx, `INSERT INTO recipes(account_id, name, description, visibility) VALUES ($1, $2, $3, $4) RETURNING id`, accountID, name, description, visibility)
	if err := row.Err(); err != nil {
		return uuid.Nil, err
	}
	var id uuid.UUID
	return id, row.Scan(&id)
}

func (m *Repository) UpdateRecipe(ctx context.Context, recipe Recipe) error {
	_, err := m.db.ExecContext(ctx, `UPDATE recipes SET name=$1, description=$2, visibility=$3 WHERE id=$4`, recipe.Name, recipe.Description, recipe.Visibility, recipe.ID)
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
