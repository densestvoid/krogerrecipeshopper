package data

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"
)

type dtx interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	PrepareNamedContext(ctx context.Context, query string) (*sqlx.NamedStmt, error)
	QueryxContext(ctx context.Context, query string, args ...interface{}) (*sqlx.Rows, error)
	SelectContext(ctx context.Context, dest any, query string, args ...interface{}) error
	Rebind(query string) string
}

func Rollback(tx *sqlx.Tx, err *error) {
	if *err == nil {
		return
	}
	if deferErr := tx.Rollback(); deferErr != nil {
		*err = errors.Join(*err, deferErr)
	}
}
