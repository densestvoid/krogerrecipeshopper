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
	LocationID      *string
}

type Session struct {
	ID        uuid.UUID
	AccountID uuid.UUID
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
	row := r.db.QueryRowContext(ctx, `SELECT id, kroger_profile_id, image_size, location_id FROM accounts WHERE kroger_profile_id = $1`, krogerProfileID)
	if err := row.Err(); err != nil {
		return Account{}, err
	}
	var account Account
	return account, row.Scan(&account.ID, &account.KrogerProfileID, &account.ImageSize, &account.LocationID)
}

func (r *Repository) GetAccountByID(ctx context.Context, id uuid.UUID) (Account, error) {
	row := r.db.QueryRowContext(ctx, `SELECT id, kroger_profile_id, image_size, location_id FROM accounts WHERE id = $1`, id)
	if err := row.Err(); err != nil {
		return Account{}, err
	}
	var account Account
	return account, row.Scan(&account.ID, &account.KrogerProfileID, &account.ImageSize, &account.LocationID)
}

func (r *Repository) DeleteAccount(ctx context.Context, id uuid.UUID) (retErr error) {
	tx, err := r.db.BeginTx(ctx, nil)
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

	// Clear ingredients
	if _, err := tx.ExecContext(ctx, `DELETE FROM ingredients USING recipes WHERE ingredients.recipe_id = recipes.id AND recipes.account_id = $1`, id); err != nil {
		return err
	}

	// Clear recipes
	if _, err := tx.ExecContext(ctx, `DELETE FROM recipes WHERE recipes.account_id = $1`, id); err != nil {
		return err
	}

	// clear favorites
	if _, err := tx.ExecContext(ctx, `DELETE FROM favorites WHERE favorites.account_id = $1`, id); err != nil {
		return err
	}

	// Clear cart
	if _, err := tx.ExecContext(ctx, `DELETE FROM cart_products WHERE cart_products.account_id = $1`, id); err != nil {
		return err
	}

	// Clear profile
	if _, err := tx.ExecContext(ctx, `DELETE FROM profiles WHERE profiles.account_id = $1`, id); err != nil {
		return err
	}

	// Clear sessions
	if _, err := tx.ExecContext(ctx, `DELETE FROM sessions WHERE sessions.account_id = $1`, id); err != nil {
		return err
	}

	// Clear account
	if _, err := tx.ExecContext(ctx, `DELETE FROM accounts WHERE accounts.id = $1`, id); err != nil {
		return err
	}

	return tx.Commit()
}

func (r *Repository) CreateSession(ctx context.Context, accountID uuid.UUID) (Session, error) {
	row := r.db.QueryRowContext(ctx, `INSERT INTO sessions(account_id) VALUES($1) RETURNING id, account_id`, accountID)
	if err := row.Err(); err != nil {
		return Session{}, err
	}
	var session Session
	return session, row.Scan(&session.ID, &session.AccountID)
}

func (r *Repository) GetSessionByID(ctx context.Context, id uuid.UUID) (Session, error) {
	row := r.db.QueryRowContext(ctx, `SELECT id, account_id FROM sessions WHERE id = $1`, id)
	if err := row.Err(); err != nil {
		return Session{}, err
	}
	var session Session
	return session, row.Scan(&session.ID, &session.AccountID)
}

func (r *Repository) DeleteSession(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM sessions WHERE id = $1`, id)
	return err
}

func (r *Repository) UpdateAccountImageSize(ctx context.Context, id uuid.UUID, imageSize string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE accounts SET image_size = $2 WHERE id = $1`, id, imageSize)
	return err
}

func (r *Repository) UpdateAccountLocationID(ctx context.Context, id uuid.UUID, locationID *string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE accounts SET location_id = $2 WHERE id = $1`, id, locationID)
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
