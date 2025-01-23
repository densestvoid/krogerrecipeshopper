package data

import (
	"context"

	"github.com/google/uuid"
)

type Account struct {
	ID              uuid.UUID
	KrogerProfileID uuid.UUID
}

type Profile struct {
	AccountID   uuid.UUID
	DisplayName string
}

const (
	ImageSizeThumbnail  = "thumbnail"
	ImageSizeSmall      = "small"
	ImageSizeMedium     = "medium"
	ImageSizeLarge      = "large"
	ImageSizeExtraLarge = "xlarge"
)

type Settings struct {
	AccountID uuid.UUID
	ImageSize string
}

func (r *Repository) CreateAccount(ctx context.Context, krogerProfileID uuid.UUID) (Account, error) {
	row := r.db.QueryRowContext(ctx, `INSERT INTO accounts(kroger_profile_id) VALUES($1) RETURNING id, kroger_profile_id`, krogerProfileID)
	if err := row.Err(); err != nil {
		return Account{}, err
	}
	var account Account
	return account, row.Scan(&account.ID, &account.KrogerProfileID)
}

func (r *Repository) GetAccountID(ctx context.Context, krogerProfileID uuid.UUID) (Account, error) {
	row := r.db.QueryRowContext(ctx, `SELECT id, kroger_profile_id FROM accounts WHERE kroger_profile_id = $1`, krogerProfileID)
	if err := row.Err(); err != nil {
		return Account{}, err
	}
	var account Account
	return account, row.Scan(&account.ID, &account.KrogerProfileID)
}
