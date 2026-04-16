package db

import "fmt"

// SearchResult represents a single search hit.
type SearchResult struct {
	NoteID  string
	Title   string
	Heading string
	Anchor  string
	Score   float64
	Snippet string
}

// FuzzySearch finds notes by LIKE matching on title and tags.
// Returns results scored by match quality.
func (d *DB) FuzzySearch(query string, limit int) ([]SearchResult, error) {
	pattern := "%" + query + "%"
	rows, err := d.conn.Query(
		`SELECT id, title, '', '',
		 CASE
		   WHEN title LIKE ? THEN 1.0
		   WHEN tags LIKE ? THEN 0.7
		   ELSE 0.5
		 END as score,
		 title
		 FROM notes
		 WHERE title LIKE ? OR tags LIKE ?
		 ORDER BY score DESC
		 LIMIT ?`,
		pattern, pattern, pattern, pattern, limit,
	)
	if err != nil {
		return nil, fmt.Errorf("fuzzy search: %w", err)
	}
	defer rows.Close()

	var results []SearchResult
	for rows.Next() {
		var r SearchResult
		if err := rows.Scan(&r.NoteID, &r.Title, &r.Heading, &r.Anchor, &r.Score, &r.Snippet); err != nil {
			return nil, fmt.Errorf("scan result: %w", err)
		}
		results = append(results, r)
	}
	return results, rows.Err()
}

// VectorSearch finds chunks by cosine similarity to the given embedding.
// This requires sqlite-vec to be loaded. If not available, returns an error.
// The embedding is a raw float32 byte slice.
func (d *DB) VectorSearch(embedding []byte, limit int) ([]SearchResult, error) {
	// sqlite-vec provides the vec_distance_cosine function.
	// We query chunks that have embeddings and rank by similarity.
	rows, err := d.conn.Query(
		`SELECT c.note_id, n.title, c.heading, c.anchor,
		 vec_distance_cosine(c.embedding, ?) as distance,
		 substr(c.content, 1, 200)
		 FROM chunks c
		 JOIN notes n ON c.note_id = n.id
		 WHERE c.embedding IS NOT NULL
		 ORDER BY distance ASC
		 LIMIT ?`,
		embedding, limit,
	)
	if err != nil {
		return nil, fmt.Errorf("vector search: %w", err)
	}
	defer rows.Close()

	var results []SearchResult
	for rows.Next() {
		var r SearchResult
		var distance float64
		if err := rows.Scan(&r.NoteID, &r.Title, &r.Heading, &r.Anchor, &distance, &r.Snippet); err != nil {
			return nil, fmt.Errorf("scan result: %w", err)
		}
		// Convert distance to similarity score (1 - distance).
		r.Score = 1.0 - distance
		results = append(results, r)
	}
	return results, rows.Err()
}
