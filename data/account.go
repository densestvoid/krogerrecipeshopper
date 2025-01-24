package data

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
)

const (
	ImageSizeThumbnail  = "thumbnail"
	ImageSizeSmall      = "small"
	ImageSizeMedium     = "medium"
	ImageSizeLarge      = "large"
	ImageSizeExtraLarge = "xlarge"
)

type Account struct {
	ID              uuid.UUID
	KrogerProfileID uuid.UUID
	ImageSize       string
}

type Profile struct {
	AccountID   uuid.UUID
	DisplayName string
}

func (r *Repository) CreateAccount(ctx context.Context, krogerProfileID uuid.UUID) (Account, error) {
	row := r.db.QueryRowContext(ctx, `INSERT INTO accounts(kroger_profile_id) VALUES($1) RETURNING id, kroger_profile_id`, krogerProfileID)
	if err := row.Err(); err != nil {
		return Account{}, err
	}
	var account Account
	return account, row.Scan(&account.ID, &account.KrogerProfileID)
}

func (r *Repository) GetAccountByKrogerProfileID(ctx context.Context, krogerProfileID uuid.UUID) (Account, error) {
	row := r.db.QueryRowContext(ctx, `SELECT id, kroger_profile_id, image_size FROM accounts WHERE kroger_profile_id = $1`, krogerProfileID)
	if err := row.Err(); err != nil {
		return Account{}, err
	}
	var account Account
	return account, row.Scan(&account.ID, &account.KrogerProfileID, &account.ImageSize)
}

func (r *Repository) GetAccountByID(ctx context.Context, id uuid.UUID) (Account, error) {
	row := r.db.QueryRowContext(ctx, `SELECT id, kroger_profile_id, image_size FROM accounts WHERE id = $1`, id)
	if err := row.Err(); err != nil {
		return Account{}, err
	}
	var account Account
	return account, row.Scan(&account.ID, &account.KrogerProfileID, &account.ImageSize)
}

func (r *Repository) UpdateAccountImageSize(ctx context.Context, id uuid.UUID, imageSize string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE accounts SET image_size = $2 WHERE id = $1`, id, imageSize)
	return err
}

func (r *Repository) CreateProfile(ctx context.Context, accountID uuid.UUID, displayName string) (Profile, error) {
	row := r.db.QueryRowContext(ctx, `
        INSERT INTO profiles (account_id, display_name)
        VALUES ($1, $2)
        RETURNING account_id, display_name
    `, accountID, displayName)
	if err := row.Err(); err != nil {
		return Profile{}, err
	}
	var profile Profile
	return profile, row.Scan(&profile.AccountID, &profile.DisplayName)
}

func (r *Repository) GetProfileByAccountID(ctx context.Context, accountID uuid.UUID) (*Profile, error) {
	row := r.db.QueryRowContext(ctx, `SELECT account_id, display_name FROM profiles WHERE account_id = $1`, accountID)
	if err := row.Err(); err != nil {
		return nil, err
	}

	var profile Profile
	if err := row.Scan(&profile.AccountID, &profile.DisplayName); errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return &profile, nil
}

func (r *Repository) ListProfiles(ctx context.Context) ([]Profile, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT account_id, display_name FROM profiles`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var profiles []Profile
	for rows.Next() {
		var profile Profile
		if err := rows.Scan(&profile.AccountID, &profile.DisplayName); err != nil {
			return nil, err
		}
		profiles = append(profiles, profile)
	}

	return profiles, nil
}

func (r *Repository) UpdateProfileDisplayName(ctx context.Context, accountID uuid.UUID, displayName string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE profiles SET display_name = $2 WHERE account_id = $1`, accountID, displayName)
	return err
}

func (r *Repository) DeleteProfile(ctx context.Context, accountID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM profiles WHERE account_id = $1`, accountID)
	return err
}
