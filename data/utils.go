package data

import (
	"context"
	"errors"

	"github.com/jmoiron/sqlx"
)

type dtx interface {
	PrepareNamedContext(ctx context.Context, query string) (*sqlx.NamedStmt, error)
}

func Rollback(tx *sqlx.Tx, err *error) {
	if *err == nil {
		return
	}
	if deferErr := tx.Rollback(); deferErr != nil {
		*err = errors.Join(*err, deferErr)
	}
}
