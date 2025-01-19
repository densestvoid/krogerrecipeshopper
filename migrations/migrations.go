package migrations

import (
	"database/sql"
	"embed"

	"github.com/pressly/goose/v3"
)

//go:embed *.sql
var migrations embed.FS

func setup() {
	goose.SetBaseFS(migrations)
	goose.SetDialect("postgres")
	goose.SetVerbose(true)
}

func Up(db *sql.DB) error {
	setup()
	return goose.Up(db, ".")
}

func Down(db *sql.DB) error {
	setup()
	return goose.Down(db, ".")
}

func Status(db *sql.DB) error {
	setup()
	return goose.Status(db, ".")
}
