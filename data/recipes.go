package data

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

const (
	VisibilityPrivate = "private"
	VisibilityFriends = "friends"
	VisibilityPublic  = "public"
)

const (
	InstructionTypeNone = "none"
	InstructionTypeText = "text"
	InstructionTypeLink = "link"
)

type Recipe struct {
	ID              uuid.UUID `db:"id"`
	AccountID       uuid.UUID `db:"account_id"`
	Name            string    `db:"name"`
	Description     string    `db:"description"`
	InstructionType string    `db:"instruction_type"`
	Instructions    string    `db:"instructions"`
	Visibility      string    `db:"visibility"`
	Favorite        bool      `db:"favorite"`
}

func (m *Repository) GetRecipe(ctx context.Context, recipeID uuid.UUID, accountID uuid.UUID) (Recipe, error) {
	var recipe Recipe
	return recipe, m.db.GetContext(ctx, &recipe, `
		SELECT
			id,
			recipes.account_id AS account_id,
			name,
			description, 
			instruction_type,
			instructions,
			visibility,
			favorites.account_id IS NOT NULL as favorite
		FROM recipes
			LEFT JOIN favorites ON favorites.recipe_id = recipes.id AND favorites.account_id = $2
		WHERE id = $1`,
		recipeID, accountID,
	)
}

type ListRecipesFilter interface {
	listRecipoesFilter(args map[string]any) string
}

type ListRecipesFilterByAccountID struct {
	AccountID uuid.UUID
}

func (f ListRecipesFilterByAccountID) listRecipoesFilter(args map[string]any) string {
	args["recipeAccountID"] = f.AccountID
	return `recipes.account_id = :recipeAccountID`
}

type ListRecipesFilterByName struct {
	Name string
}

func (f ListRecipesFilterByName) listRecipoesFilter(args map[string]any) string {
	args["recipeName"] = fmt.Sprintf("%%%s%%", f.Name)
	return `name ILIKE :recipeName`
}

type ListRecipesFilterByFavorites struct{}

func (f ListRecipesFilterByFavorites) listRecipoesFilter(args map[string]any) string {
	return `favorites.account_id IS NOT NULL`
}

type ListRecipesFilterByVisibilities struct {
	Visibilities []string
}

func (f ListRecipesFilterByVisibilities) listRecipoesFilter(args map[string]any) string {
	if len(f.Visibilities) == 0 {
		return `false`
	}
	args["recipeVisibilities"] = f.Visibilities
	return `visibility = ANY(:recipeVisibilities)`
}

type ListRecipesOrderBy struct {
	Field     string
	Direction string
}

func (m *Repository) ListRecipes(ctx context.Context, accountID uuid.UUID, filters []ListRecipesFilter, orderBys []ListRecipesOrderBy) ([]Recipe, error) {
	query := `
		SELECT
			id,
			recipes.account_id,
			name,
			description, 
			instruction_type,
			instructions,
			visibility,
			favorites.account_id IS NOT NULL as favorite
		FROM recipes
			LEFT JOIN favorites ON favorites.recipe_id = recipes.id AND favorites.account_id = :accountID
		WHERE (recipes.account_id = :accountID OR visibility = 'public')
	`
	namedArgs := map[string]any{"accountID": accountID}
	if len(filters) > 0 {
		filterStrings := []string{}
		for _, filter := range filters {
			filterString := filter.listRecipoesFilter(namedArgs)
			filterStrings = append(filterStrings, filterString)
		}
		query += " AND " + strings.Join(filterStrings, " AND ")
	}
	if len(orderBys) > 0 {
		query += " ORDER BY "
		orderStrings := []string{}
		for _, orderBy := range orderBys {
			orderStrings = append(orderStrings, fmt.Sprintf("%s %s", orderBy.Field, orderBy.Direction))
		}
		query += strings.Join(orderStrings, ", ")
	}

	namedQuery, err := m.db.PrepareNamedContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer namedQuery.Close()

	var recipes = []Recipe{}
	return recipes, namedQuery.Select(&recipes, namedArgs)
}

func (m *Repository) CreateRecipe(ctx context.Context, accountID uuid.UUID, name, description, instructionType, instructions, visibility string) (uuid.UUID, error) {
	namedQuery, err := m.db.PrepareNamedContext(ctx, `
		INSERT INTO recipes(
			account_id,
			name,
			description,
			instruction_type,
			instructions,
			visibility
		) VALUES (
			:accountID,
			:name,
			:description,
			:instructionType,
			:instructions,
			:visibility
		) RETURNING id
	`)
	if err != nil {
		return uuid.Nil, err
	}

	var id uuid.UUID
	return id, namedQuery.GetContext(ctx, &id, map[string]any{
		"accountID":       accountID,
		"name":            name,
		"description":     description,
		"instructionType": instructionType,
		"instructions":    instructions,
		"visibility":      visibility,
	})
}

func (m *Repository) UpdateRecipe(ctx context.Context, recipe Recipe) error {
	namedStmt, err := m.db.PrepareNamedContext(ctx, `
		UPDATE recipes
		SET
			name=:name,
			description=:description,
			instruction_type=:instructionType,
			instructions=:instructions,
			visibility=:visibility
		WHERE id=:id
	`)
	if err != nil {
		return err
	}
	_, err = namedStmt.ExecContext(ctx, map[string]any{
		"name":            recipe.Name,
		"description":     recipe.Description,
		"instructionType": recipe.InstructionType,
		"instructions":    recipe.Instructions,
		"visibility":      recipe.Visibility,
		"id":              recipe.ID,
	})
	return err
}

func (m *Repository) FavoriteRecipe(ctx context.Context, recipeID, accountID uuid.UUID) error {
	_, err := m.db.ExecContext(ctx, `INSERT INTO favorites(recipe_id, account_id) VALUES ($1, $2)`, recipeID, accountID)
	return err
}

func (m *Repository) UnfavoriteRecipe(ctx context.Context, recipeID, accountID uuid.UUID) error {
	_, err := m.db.ExecContext(ctx, `DELETE FROM favorites WHERE recipe_id = $1 AND account_id = $2`, recipeID, accountID)
	return err
}

func (m *Repository) DeleteRecipe(ctx context.Context, id uuid.UUID) (retErr error) {
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	defer func() {
		if retErr == nil {
			return
		}
		if deferErr := tx.Rollback(); deferErr != nil {
			retErr = errors.Join(retErr, deferErr)
		}
	}()

	if _, err := tx.ExecContext(ctx, `DELETE FROM ingredients where recipe_id = $1`, id); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM recipes WHERE id = $1`, id); err != nil {
		return err
	}
	return tx.Commit()
}
