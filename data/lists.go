package data

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

type List struct {
	ID          uuid.UUID `db:"id"`
	AccountID   uuid.UUID `db:"account_id"`
	Name        string    `db:"name"`
	Description string    `db:"description"`
}

func (r *Repository) GetList(ctx context.Context, listID uuid.UUID) (List, error) {
	var list List
	return list, r.db.GetContext(ctx, &list, `SELECT id, account_id, name, description FROM lists WHERE id = $1`, listID)
}

type ListListsFilter interface {
	listListsFilter(args map[string]any) string
}
type ListListsFilterByName struct {
	Name string
}

func (f ListListsFilterByName) listListsFilter(args map[string]any) string {
	args["recipeName"] = fmt.Sprintf("%%%s%%", f.Name)
	return `name ILIKE :recipeName`
}

type ListListsOrderBy struct {
	Field     string
	Direction string
}

func (r *Repository) ListLists(ctx context.Context, accountID uuid.UUID, filters []ListListsFilter, orderBys []ListListsOrderBy) ([]List, error) {
	query := `
		SELECT
			id,
			account_id,
            name,
            description
		FROM lists
			LEFT JOIN recipes ON recipes.list_id = lists.id
		WHERE account_id = :accountID AND recipes.list_id IS NULL
	`
	namedArgs := map[string]any{"accountID": accountID}
	if len(filters) > 0 {
		filterStrings := []string{}
		for _, filter := range filters {
			filterString := filter.listListsFilter(namedArgs)
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

	var lists = []List{}
	return lists, namedQuery.Select(&lists, namedArgs)
}

func (r *Repository) CreateList(ctx context.Context, accountID uuid.UUID, name, description string) (uuid.UUID, error) {
	return r.createList(ctx, r.db, accountID, name, description)
}

func (r *Repository) createList(ctx context.Context, dtx dtx, accountID uuid.UUID, name, description string) (uuid.UUID, error) {
	namedQuery, err := dtx.PrepareNamedContext(ctx, `
		INSERT INTO lists(
			account_id,
			name,
			description
		) VALUES (
			:accountID,
			:name,
			:description
		) RETURNING id
	`)
	if err != nil {
		return uuid.Nil, err
	}

	var id uuid.UUID
	return id, namedQuery.GetContext(ctx, &id, map[string]any{
		"accountID":   accountID,
		"name":        name,
		"description": description,
	})
}

func (r *Repository) UpdateList(ctx context.Context, list List) error {
	return r.updateList(ctx, r.db, list)
}

func (r *Repository) updateList(ctx context.Context, dtx dtx, list List) error {
	namedQuery, err := dtx.PrepareNamedContext(ctx, `
        UPDATE lists
        SET name = :name, description = :description
        WHERE id = :id
    `)
	if err != nil {
		return err
	}

	_, err = namedQuery.ExecContext(ctx, map[string]any{
		"id":          list.ID,
		"name":        list.Name,
		"description": list.Description,
	})
	return err
}

func (r *Repository) DeleteList(ctx context.Context, id uuid.UUID) (retErr error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer Rollback(tx, &retErr)

	if _, err := tx.ExecContext(ctx, `DELETE FROM ingredients where list_id = $1`, id); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM lists WHERE id = $1`, id); err != nil {
		return err
	}
	return tx.Commit()
}
