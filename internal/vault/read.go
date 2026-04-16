package vault

import (
	"fmt"
	"os"
	"path/filepath"
)

// ReadResult is what vault.read returns.
type ReadResult struct {
	ID          string
	Title       string
	Tags        []string
	Body        string
	RelatesTo   []string
	CreatedAt   string
	UpdatedAt   string
	RawContent  string // full file content including frontmatter
}

// Read retrieves a note by ID (slug).
func (s *Service) Read(id string) (*ReadResult, error) {
	note, err := s.db.GetNote(id)
	if err != nil {
		return nil, err
	}
	if note == nil {
		return nil, fmt.Errorf("note not found: %s", id)
	}

	fullPath := filepath.Join(s.vaultDir, note.FilePath)
	raw, err := os.ReadFile(fullPath)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}

	fm, body, err := ParseFrontmatter(string(raw))
	if err != nil {
		return nil, err
	}

	return &ReadResult{
		ID:         note.ID,
		Title:      fm.Title,
		Tags:       fm.Tags,
		Body:       body,
		RelatesTo:  fm.RelatesTo,
		CreatedAt:  fm.CreatedAt,
		UpdatedAt:  fm.UpdatedAt,
		RawContent: string(raw),
	}, nil
}
