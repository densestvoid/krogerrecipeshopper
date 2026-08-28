package data

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/lib/pq"
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

const recipeTagsJoin = `
	LEFT JOIN (
		SELECT recipe_list_id, COALESCE(ARRAY_AGG(name ORDER BY name), '{}') AS tags
		FROM recipe_tags
			INNER JOIN tags ON tags.id = recipe_tags.tag_id
		GROUP BY recipe_list_id
	) AS tags ON tags.recipe_list_id = recipes.list_id
`

type Recipe struct {
	ListID          uuid.UUID `db:"list_id"`
	AccountID       uuid.UUID `db:"account_id"`
	Name            string    `db:"name"`
	Description     string    `db:"description"`
	InstructionType string    `db:"instruction_type"`
	Instructions    string    `db:"instructions"`
	Visibility      string    `db:"visibility"`
	Favorite        bool      `db:"favorite"`
	Tags            []string
}

type listRecipe struct {
	ListID          uuid.UUID      `db:"list_id"`
	AccountID       uuid.UUID      `db:"account_id"`
	Name            string         `db:"name"`
	Description     string         `db:"description"`
	InstructionType string         `db:"instruction_type"`
	Instructions    string         `db:"instructions"`
	Visibility      string         `db:"visibility"`
	Favorite        bool           `db:"favorite"`
	Tags            pq.StringArray `db:"tags"`
}

func (r listRecipe) recipe() Recipe {
	return Recipe{
		ListID:          r.ListID,
		AccountID:       r.AccountID,
		Name:            r.Name,
		Description:     r.Description,
		InstructionType: r.InstructionType,
		Instructions:    r.Instructions,
		Visibility:      r.Visibility,
		Favorite:        r.Favorite,
		Tags:            []string(r.Tags),
	}
}

func (r *Repository) GetRecipe(ctx context.Context, listID uuid.UUID, accountID uuid.UUID) (Recipe, error) {
	namedQuery, err := r.db.PrepareNamedContext(ctx, `
		SELECT
			recipes.*,
			favorites.account_id IS NOT NULL as favorite,
			COALESCE(tags.tags, '{}') as tags
		FROM recipe_list_view AS recipes
			LEFT JOIN favorites ON favorites.list_id = recipes.list_id AND favorites.account_id = :accountID
		`+recipeTagsJoin+`
		WHERE recipes.list_id = :listID
	`)
	if err != nil {
		return Recipe{}, err
	}
	defer namedQuery.Close()

	var row listRecipe
	if err := namedQuery.GetContext(ctx, &row, map[string]any{
		"listID":    listID,
		"accountID": accountID,
	}); err != nil {
		return Recipe{}, err
	}

	return row.recipe(), nil
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

type ListRecipesFilterByTags struct {
	Tags []string
}

func (f ListRecipesFilterByTags) listRecipoesFilter(args map[string]any) string {
	if len(f.Tags) == 0 {
		return `false`
	}
	args["recipeTags"] = f.Tags
	return `COALESCE(tags.tags, '{}') @> :recipeTags`
}

type ListRecipesOrderBy struct {
	Field     string
	Direction string
}

func (r *Repository) ListRecipes(ctx context.Context, accountID uuid.UUID, filters []ListRecipesFilter, orderBys []ListRecipesOrderBy) ([]Recipe, error) {
	query := `
		SELECT
			recipes.*,
			favorites.account_id IS NOT NULL as favorite,
			COALESCE(tags.tags, '{}') as tags
		FROM recipe_list_view AS recipes
			LEFT JOIN favorites ON favorites.list_id = recipes.list_id AND favorites.account_id = :accountID
		` + recipeTagsJoin + `
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

	namedQuery, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer namedQuery.Close()

	var rows []listRecipe
	if err := namedQuery.Select(&rows, namedArgs); err != nil {
		return nil, err
	}

	recipes := make([]Recipe, len(rows))
	for i, row := range rows {
		recipes[i] = row.recipe()
	}
	return recipes, nil
}

func (r *Repository) CreateRecipe(ctx context.Context, accountID uuid.UUID, name, description, instructionType, instructions, visibility string, tags []string) (listID uuid.UUID, retErr error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return uuid.Nil, err
	}
	defer Rollback(tx, &retErr)

	listID, err = r.createList(ctx, tx, accountID, name, description)
	if err != nil {
		return uuid.Nil, err
	}

	namedQuery, err := tx.PrepareNamedContext(ctx, `
		INSERT INTO recipes(
		    list_id,
			instruction_type,
			instructions,
			visibility
		) VALUES (
			:listID,
			:instructionType,
			:instructions,
			:visibility
		)
	`)
	if err != nil {
		return uuid.Nil, err
	}

	if _, err := namedQuery.ExecContext(ctx, map[string]any{
		"listID":          listID,
		"instructionType": instructionType,
		"instructions":    instructions,
		"visibility":      visibility,
	}); err != nil {
		return uuid.Nil, err
	}

	if err := r.setRecipeTags(ctx, tx, listID, tags); err != nil {
		return uuid.Nil, err
	}

	return listID, tx.Commit()
}

func (r *Repository) UpdateRecipe(ctx context.Context, recipe Recipe) (retErr error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer Rollback(tx, &retErr)

	if err := r.updateList(ctx, tx, List{
		ID:          recipe.ListID,
		AccountID:   recipe.AccountID,
		Name:        recipe.Name,
		Description: recipe.Description,
	}); err != nil {
		return err
	}

	namedQuery, err := r.db.PrepareNamedContext(ctx, `
		UPDATE recipes
		SET instruction_type = :instructionType, instructions = :instructions, visibility = :visibility
		WHERE list_id = :listID
	`)
	if err != nil {
		return err
	}

	if _, err := namedQuery.ExecContext(ctx, map[string]any{
		"listID":          recipe.ListID,
		"instructionType": recipe.InstructionType,
		"instructions":    recipe.Instructions,
		"visibility":      recipe.Visibility,
	}); err != nil {
		return err
	}

	if err := r.setRecipeTags(ctx, tx, recipe.ListID, recipe.Tags); err != nil {
		return err
	}

	return tx.Commit()
}

func (m *Repository) FavoriteRecipe(ctx context.Context, listID, accountID uuid.UUID) error {
	_, err := m.db.ExecContext(ctx, `INSERT INTO favorites(list_id, account_id) VALUES ($1, $2)`, listID, accountID)
	return err
}

func (m *Repository) UnfavoriteRecipe(ctx context.Context, listID, accountID uuid.UUID) error {
	_, err := m.db.ExecContext(ctx, `DELETE FROM favorites WHERE list_id = $1 AND account_id = $2`, listID, accountID)
	return err
}

func (m *Repository) DeleteRecipe(ctx context.Context, listID uuid.UUID) (retErr error) {
	tx, err := m.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer Rollback(tx, &retErr)

	if _, err := tx.ExecContext(ctx, `DELETE FROM recipe_tags WHERE recipe_list_id = $1`, listID); err != nil {
		return err
	}
	if err := m.deleteOrphanTags(ctx, tx); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM favorites WHERE list_id = $1`, listID); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM ingredients where list_id = $1`, listID); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM recipes WHERE list_id = $1`, listID); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM lists WHERE id = $1`, listID); err != nil {
		return err
	}
	return tx.Commit()
}
