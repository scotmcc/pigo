package main

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

// openDBForStats opens the SQLite database read-only for quick stats.
// Separate from the full db package to avoid pulling in migration logic.
func openDBForStats(path string) (*sql.DB, error) {
	return sql.Open("sqlite3", path+"?mode=ro")
}
