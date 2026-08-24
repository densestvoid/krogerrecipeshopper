package data

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

const MaxTagNameLength = 32

var ErrTagNameTooLong = errors.New("tag name exceeds maximum length")

func NormalizeTagNames(tagNames []string) ([]string, error) {
	seen := make(map[string]struct{})
	filtered := []string{}
	for _, name := range tagNames {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		name = strings.ToLower(name)
		if len(name) > MaxTagNameLength {
			return nil, fmt.Errorf("%w (%d characters max): %q", ErrTagNameTooLong, MaxTagNameLength, name)
		}
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}
		filtered = append(filtered, name)
	}
	return filtered, nil
}

func (r *Repository) ListTags(ctx context.Context, stub string, excludes []string) ([]string, error) {
	excludes, err := NormalizeTagNames(excludes)
	if err != nil {
		return nil, err
	}

	stub = strings.ToLower(strings.TrimSpace(stub))

	var query string
	var args []any

	if len(excludes) > 0 {
		query = `
			SELECT tags.name
			FROM tags
				INNER JOIN recipe_tags ON recipe_tags.tag_id = tags.id
			WHERE tags.name ILIKE ? AND tags.name NOT IN (?)
			GROUP BY tags.name
			ORDER BY COUNT(recipe_tags.recipe_list_id) DESC
			LIMIT 10
		`
		query, args, err = sqlx.In(query, "%"+stub+"%", excludes)
		if err != nil {
			return nil, err
		}
	} else {
		query = `
			SELECT tags.name
			FROM tags
				INNER JOIN recipe_tags ON recipe_tags.tag_id = tags.id
			WHERE tags.name ILIKE ?
			GROUP BY tags.name
			ORDER BY COUNT(recipe_tags.recipe_list_id) DESC
			LIMIT 10
		`
		args = []any{"%" + stub + "%"}
	}

	query = r.db.Rebind(query)
	var tags []string
	return tags, r.db.SelectContext(ctx, &tags, query, args...)
}

func (r *Repository) listRecipeTags(ctx context.Context, recipeListID uuid.UUID) ([]string, error) {
	var tags []string
	return tags, r.db.SelectContext(ctx, &tags, `
		SELECT name
		FROM tags
			INNER JOIN recipe_tags ON recipe_tags.tag_id = tags.id
		WHERE recipe_list_id = $1
	`, recipeListID)
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
	normalized, err := NormalizeTagNames(tagNames)
	if err != nil {
		return err
	}
	tagNames = normalized

	if len(tagNames) == 0 {
		if _, err := dtx.ExecContext(ctx, `DELETE FROM recipe_tags WHERE recipe_list_id = $1`, recipeListID); err != nil {
			return err
		}
		return r.deleteOrphanTags(ctx, dtx)
	}

	var tagIDs []uuid.UUID
	tags := []map[string]any{}
	for _, name := range tagNames {
		tags = append(tags, map[string]any{
			"name": name,
		})
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

	query, args, err = sqlx.Named(`DELETE FROM recipe_tags WHERE recipe_list_id = :recipeListID AND tag_id NOT IN (:tagIDs)`, map[string]any{
		"recipeListID": recipeListID,
		"tagIDs":       tagIDs,
	})
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

	query, args, err = sqlx.Named(`
		INSERT INTO recipe_tags (recipe_list_id, tag_id)
		SELECT :recipeListID, id
		FROM tags
		WHERE id IN (:tagIDs) AND id NOT IN (
			SELECT tag_id
			FROM recipe_tags
			WHERE recipe_list_id = :recipeListID
		)
	`, map[string]any{
		"recipeListID": recipeListID,
		"tagIDs":       tagIDs,
	})
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
