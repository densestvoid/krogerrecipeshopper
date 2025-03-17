package data

import (
	"context"

	"github.com/google/uuid"
)

type List struct {
	ID          uuid.UUID `db:"id"`
	AccountID   uuid.UUID `db:"account_id"`
	Name        string    `db:"name"`
	Description string    `db:"description"`
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
