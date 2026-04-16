package vault

import (
	"github.com/scotmcc/pigo/internal/db"
)

// SearchResult wraps the results with optional warnings about degraded operation.
type SearchResult struct {
	Results  []db.SearchResult
	Warnings []string
}

// Search finds notes by combining fuzzy title/tag matching with semantic search.
// Results are merged and deduplicated, with the best score for each note kept.
// If Ollama is unavailable, fuzzy-only results are returned with a warning.
// sqlite-vec is statically bundled so vector search is always available.
func (s *Service) Search(query string, limit int) (*SearchResult, error) {
	if limit <= 0 {
		limit = 10
	}

	var warnings []string

	// Run fuzzy search (always works, no embedding needed).
	fuzzy, err := s.db.FuzzySearch(query, limit)
	if err != nil {
		return nil, err
	}

	// Try semantic search (requires Ollama for the query embedding).
	var semantic []db.SearchResult
	embedding, embedErr := s.ollama.Embed(query)
	if embedErr != nil {
		warnings = append(warnings, "semantic search unavailable (Ollama not running) — showing fuzzy results only")
	} else {
		var vecErr error
		semantic, vecErr = s.db.VectorSearch(embedding, limit)
		if vecErr != nil {
			warnings = append(warnings, "vector search failed — showing fuzzy results only")
		}
	}

	return &SearchResult{
		Results:  mergeResults(fuzzy, semantic, limit),
		Warnings: warnings,
	}, nil
}

// mergeResults combines fuzzy and semantic results, keeping the best score per note.
func mergeResults(fuzzy, semantic []db.SearchResult, limit int) []db.SearchResult {
	seen := make(map[string]*db.SearchResult)

	for i := range fuzzy {
		r := &fuzzy[i]
		seen[r.NoteID] = r
	}

	for i := range semantic {
		r := &semantic[i]
		if existing, ok := seen[r.NoteID]; ok {
			if r.Score > existing.Score {
				seen[r.NoteID] = r
			}
		} else {
			seen[r.NoteID] = r
		}
	}

	results := make([]db.SearchResult, 0, len(seen))
	for _, r := range seen {
		results = append(results, *r)
	}

	// Sort by score descending.
	for i := 1; i < len(results); i++ {
		for j := i; j > 0 && results[j].Score > results[j-1].Score; j-- {
			results[j], results[j-1] = results[j-1], results[j]
		}
	}

	if len(results) > limit {
		results = results[:limit]
	}

	return results
}
