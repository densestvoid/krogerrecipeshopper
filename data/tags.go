package data

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

const MaxTagNameLength = 32

func (r *Repository) ListTags(ctx context.Context, stub string, excludes []string) ([]string, error) {
	if excludes == nil {
		excludes = []string{}
	}

	stub = strings.ToLower(strings.TrimSpace(stub))

	var tags []string
	return tags, r.db.SelectContext(ctx, &tags, `
		SELECT tags.name
		FROM tags
			INNER JOIN recipe_tags ON recipe_tags.tag_id = tags.id
		WHERE tags.name ILIKE $1
			AND NOT (tags.name = ANY($2))
		GROUP BY tags.name
		ORDER BY COUNT(recipe_tags.recipe_list_id) DESC
		LIMIT 10
	`, "%"+stub+"%", excludes)
}

func (r *Repository) deleteOrphanTags(ctx context.Context, dtx dtx) error {
	_, err := dtx.ExecContext(ctx, `
		DELETE FROM tags
		WHERE id NOT IN (
			SELECT tag_id
			FROM recipe_tags
		)
	`)
	return err
}

func (r *Repository) setRecipeTags(ctx context.Context, dtx dtx, recipeListID uuid.UUID, tagNames []string) error {
	if _, err := dtx.ExecContext(ctx, `DELETE FROM recipe_tags WHERE recipe_list_id = $1`, recipeListID); err != nil {
		return err
	}

	if len(tagNames) == 0 {
		return r.deleteOrphanTags(ctx, dtx)
	}

	tags := make([]map[string]any, len(tagNames))
	for i, name := range tagNames {
		tags[i] = map[string]any{"name": name}
	}

	query, args, err := sqlx.Named(`INSERT INTO tags (name) VALUES (:name) ON CONFLICT ((lower(name))) DO NOTHING`, tags)
	if err != nil {
		return err
	}
	query, args, err = sqlx.In(query, args...)
	if err != nil {
		return err
	}
	query = dtx.Rebind(query)
	if _, err := dtx.ExecContext(ctx, query, args...); err != nil {
		return err
	}

	var tagIDs []uuid.UUID
	query, args, err = sqlx.Named(`SELECT id FROM tags WHERE name IN (:tagNames)`, map[string]any{
		"tagNames": tagNames,
	})
	if err != nil {
		return err
	}
	query, args, err = sqlx.In(query, args...)
	if err != nil {
		return err
	}
	query = dtx.Rebind(query)
	if err := dtx.SelectContext(ctx, &tagIDs, query, args...); err != nil {
		return err
	}

	recipeTags := make([]map[string]any, len(tagIDs))
	for i, tagID := range tagIDs {
		recipeTags[i] = map[string]any{
			"recipeListID": recipeListID,
			"tagID":        tagID,
		}
	}

	query, args, err = sqlx.Named(`INSERT INTO recipe_tags (recipe_list_id, tag_id) VALUES (:recipeListID, :tagID)`, recipeTags)
	if err != nil {
		return err
	}
	query, args, err = sqlx.In(query, args...)
	if err != nil {
		return err
	}
	query = dtx.Rebind(query)
	if _, err := dtx.ExecContext(ctx, query, args...); err != nil {
		return err
	}

	return r.deleteOrphanTags(ctx, dtx)
}
