package db

import "fmt"

// Chunk represents a row in the chunks table.
type Chunk struct {
	ID        string
	NoteID    string
	Heading   string // may be empty for intro chunk
	Anchor    string // #heading-slug for deep linking
	Content   string
	Embedding []byte // float32 vector as raw bytes
}

// InsertChunks inserts multiple chunks in a single transaction.
func (d *DB) InsertChunks(chunks []Chunk) error {
	tx, err := d.conn.Begin()
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(
		`INSERT INTO chunks (id, note_id, heading, anchor, content, embedding)
		 VALUES (?, ?, ?, ?, ?, ?)`,
	)
	if err != nil {
		return fmt.Errorf("prepare insert: %w", err)
	}
	defer stmt.Close()

	for _, c := range chunks {
		if _, err := stmt.Exec(c.ID, c.NoteID, c.Heading, c.Anchor, c.Content, c.Embedding); err != nil {
			return fmt.Errorf("insert chunk: %w", err)
		}
	}

	return tx.Commit()
}

// DeleteChunksByNoteID removes all chunks for a given note.
func (d *DB) DeleteChunksByNoteID(noteID string) error {
	_, err := d.conn.Exec(`DELETE FROM chunks WHERE note_id = ?`, noteID)
	if err != nil {
		return fmt.Errorf("delete chunks: %w", err)
	}
	return nil
}

// GetChunksByNoteID returns all chunks for a note, ordered by rowid.
func (d *DB) GetChunksByNoteID(noteID string) ([]Chunk, error) {
	rows, err := d.conn.Query(
		`SELECT id, note_id, heading, anchor, content, embedding
		 FROM chunks WHERE note_id = ? ORDER BY rowid`, noteID,
	)
	if err != nil {
		return nil, fmt.Errorf("get chunks: %w", err)
	}
	defer rows.Close()

	var chunks []Chunk
	for rows.Next() {
		var c Chunk
		if err := rows.Scan(&c.ID, &c.NoteID, &c.Heading, &c.Anchor, &c.Content, &c.Embedding); err != nil {
			return nil, fmt.Errorf("scan chunk: %w", err)
		}
		chunks = append(chunks, c)
	}
	return chunks, rows.Err()
}
