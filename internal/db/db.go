// Package db provides the SQLite connection and schema management.
// This is a layer-1 package — it talks to SQLite and nothing else.
package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	sqlite_vec "github.com/asg017/sqlite-vec-go-bindings/cgo"
	_ "github.com/mattn/go-sqlite3" // SQLite driver, imported for side effects
)

// sqlite-vec is auto-registered at package load so every sqlite3 connection
// opened through database/sql has vec_distance_cosine and friends available.
// Bundling it this way means semantic search works out of the box — users
// never see or install a separate extension file.
func init() {
	sqlite_vec.Auto()
}

// DB wraps a sql.DB connection with pigo-specific operations.
type DB struct {
	conn *sql.DB
}

// Open creates or opens a SQLite database at the given path.
// It creates parent directories if needed and sets pragmas for performance.
func Open(path string) (*DB, error) {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("create db directory: %w", err)
	}

	conn, err := sql.Open("sqlite3", path+"?_journal_mode=WAL&_busy_timeout=5000")
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	if err := conn.Ping(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}

	return &DB{conn: conn}, nil
}

// Close closes the database connection.
func (d *DB) Close() error {
	return d.conn.Close()
}

// Conn returns the underlying sql.DB for direct access when needed.
func (d *DB) Conn() *sql.DB {
	return d.conn
}
