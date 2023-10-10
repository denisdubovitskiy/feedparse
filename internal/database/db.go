package database

import (
	"context"
	"database/sql"
	_ "embed"

	_ "github.com/mattn/go-sqlite3"
)

func Open(name string) (*sql.DB, error) {
	return sql.Open("sqlite3", name)
}

//go:embed migration.sql
var migration string

func Migrate(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, migration)
	return err
}
