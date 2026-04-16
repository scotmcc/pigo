package db

import "fmt"

// Migrate runs schema migrations. It creates tables if they don't exist.
// Migrations are idempotent — safe to run on every startup.
func (d *DB) Migrate() error {
	for _, stmt := range migrations {
		if _, err := d.conn.Exec(stmt); err != nil {
			return fmt.Errorf("migration failed: %w", err)
		}
	}
	return nil
}

// migrations are the schema statements, run in order.
// Each uses IF NOT EXISTS so they're safe to re-run.
var migrations = []string{
	`CREATE TABLE IF NOT EXISTS notes (
		id        TEXT PRIMARY KEY,
		file_path TEXT NOT NULL,
		title     TEXT NOT NULL,
		tags      TEXT NOT NULL DEFAULT '[]',
		created_at DATETIME NOT NULL DEFAULT (datetime('now')),
		updated_at DATETIME NOT NULL DEFAULT (datetime('now'))
	)`,

	`CREATE TABLE IF NOT EXISTS chunks (
		id       TEXT PRIMARY KEY,
		note_id  TEXT NOT NULL,
		heading  TEXT,
		anchor   TEXT,
		content  TEXT NOT NULL,
		embedding BLOB,
		FOREIGN KEY (note_id) REFERENCES notes(id) ON DELETE CASCADE
	)`,

	`CREATE INDEX IF NOT EXISTS idx_chunks_note_id ON chunks(note_id)`,

	`CREATE TABLE IF NOT EXISTS facts (
		id         TEXT PRIMARY KEY,
		source_id  TEXT NOT NULL,
		content    TEXT NOT NULL,
		topic      TEXT,
		importance INTEGER DEFAULT 0,
		entities   TEXT NOT NULL DEFAULT '[]',
		created_at DATETIME NOT NULL DEFAULT (datetime('now')),
		FOREIGN KEY (source_id) REFERENCES notes(id) ON DELETE CASCADE
	)`,

	`CREATE INDEX IF NOT EXISTS idx_facts_source_id ON facts(source_id)`,
	`CREATE INDEX IF NOT EXISTS idx_facts_topic ON facts(topic)`,
}
