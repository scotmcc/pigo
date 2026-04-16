package db

import (
	"database/sql"
	"fmt"
	"time"
)

// Note represents a row in the notes table.
type Note struct {
	ID        string
	FilePath  string
	Title     string
	Tags      string // JSON array
	CreatedAt time.Time
	UpdatedAt time.Time
}

// InsertNote creates a new note record.
func (d *DB) InsertNote(n Note) error {
	_, err := d.conn.Exec(
		`INSERT INTO notes (id, file_path, title, tags, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		n.ID, n.FilePath, n.Title, n.Tags, n.CreatedAt, n.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert note: %w", err)
	}
	return nil
}

// UpdateNote updates an existing note's title, tags, and updated_at.
func (d *DB) UpdateNote(n Note) error {
	_, err := d.conn.Exec(
		`UPDATE notes SET title = ?, tags = ?, updated_at = ? WHERE id = ?`,
		n.Title, n.Tags, n.UpdatedAt, n.ID,
	)
	if err != nil {
		return fmt.Errorf("update note: %w", err)
	}
	return nil
}

// GetNote retrieves a note by ID.
func (d *DB) GetNote(id string) (*Note, error) {
	row := d.conn.QueryRow(
		`SELECT id, file_path, title, tags, created_at, updated_at FROM notes WHERE id = ?`, id,
	)
	return scanNote(row)
}

// GetNoteByPath retrieves a note by file path.
func (d *DB) GetNoteByPath(path string) (*Note, error) {
	row := d.conn.QueryRow(
		`SELECT id, file_path, title, tags, created_at, updated_at FROM notes WHERE file_path = ?`, path,
	)
	return scanNote(row)
}

// ListNotes returns all notes ordered by updated_at descending.
func (d *DB) ListNotes() ([]Note, error) {
	rows, err := d.conn.Query(
		`SELECT id, file_path, title, tags, created_at, updated_at FROM notes ORDER BY updated_at DESC`,
	)
	if err != nil {
		return nil, fmt.Errorf("list notes: %w", err)
	}
	defer rows.Close()

	var notes []Note
	for rows.Next() {
		var n Note
		if err := rows.Scan(&n.ID, &n.FilePath, &n.Title, &n.Tags, &n.CreatedAt, &n.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan note: %w", err)
		}
		notes = append(notes, n)
	}
	return notes, rows.Err()
}

// DeleteNote removes a note by ID. Chunks are cascade-deleted.
func (d *DB) DeleteNote(id string) error {
	_, err := d.conn.Exec(`DELETE FROM notes WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete note: %w", err)
	}
	return nil
}

// scanNote scans a single note row. Returns nil, nil if not found.
func scanNote(row *sql.Row) (*Note, error) {
	var n Note
	err := row.Scan(&n.ID, &n.FilePath, &n.Title, &n.Tags, &n.CreatedAt, &n.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("scan note: %w", err)
	}
	return &n, nil
}
