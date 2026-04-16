package vault

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// EditInput is what the caller provides to update a note.
// Only non-nil fields are applied — this allows partial updates.
type EditInput struct {
	ID   string
	Body *string  // new body, or nil to keep existing
	Tags []string // new tags, or nil to keep existing
}

// Edit updates an existing note in the vault.
// It re-renders the file, re-indexes chunks, and commits to git.
func (s *Service) Edit(input EditInput) error {
	// Read the current note.
	note, err := s.db.GetNote(input.ID)
	if err != nil {
		return err
	}
	if note == nil {
		return fmt.Errorf("note not found: %s", input.ID)
	}

	fullPath := filepath.Join(s.vaultDir, note.FilePath)
	raw, err := os.ReadFile(fullPath)
	if err != nil {
		return fmt.Errorf("read file: %w", err)
	}

	fm, body, err := ParseFrontmatter(string(raw))
	if err != nil {
		return err
	}

	// Apply changes.
	if input.Body != nil {
		body = *input.Body
	}
	if input.Tags != nil {
		fm.Tags = input.Tags
	}
	fm.UpdatedAt = time.Now().UTC().Format(time.RFC3339)

	// Re-render and write.
	content, err := RenderNote(fm, body)
	if err != nil {
		return err
	}
	if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("write file: %w", err)
	}

	// Update index.
	tagsJSON, _ := json.Marshal(fm.Tags)
	note.Tags = string(tagsJSON)
	note.UpdatedAt = time.Now().UTC()
	if err := s.db.UpdateNote(*note); err != nil {
		return err
	}

	// Re-index chunks.
	if err := s.indexChunks(input.ID, body); err != nil {
		return fmt.Errorf("re-index chunks: %w", err)
	}

	// Git commit.
	if err := s.git.CommitFile(note.FilePath, fmt.Sprintf("vault: %s [edit]", fm.Title)); err != nil {
		return fmt.Errorf("git commit: %w", err)
	}

	return nil
}
