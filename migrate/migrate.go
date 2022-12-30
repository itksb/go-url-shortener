package migrate

import (
	"database/sql"
	"embed"
	"github.com/pressly/goose/v3"
	"io/fs"
)

// Migrations - virtual file system
//
//go:embed migrations/*.sql
var Migrations embed.FS

// Migrate - .
func Migrate(dsn string, path fs.FS) error {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return err
	}

	defer db.Close()

	goose.SetBaseFS(path)
	return goose.Up(db, "migrations")
}
