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
	if err := goose.SetDialect("postgres"); err != nil {
		panic(err)
	}
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
