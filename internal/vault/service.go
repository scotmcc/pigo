// Package vault implements the four vault operations: read, write, edit, search.
// This is a layer-2 package — it composes layer-1 calls (db, git, ollama).
package vault

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/scotmcc/pigo/internal/db"
	"github.com/scotmcc/pigo/internal/git"
	"github.com/scotmcc/pigo/internal/ollama"
)

// Service is the vault business logic layer.
// It holds references to all three layer-1 dependencies.
type Service struct {
	db       *db.DB
	git      *git.Repo
	ollama   *ollama.Client
	vaultDir string // root directory where markdown files live
}

// NewService creates a vault service with its dependencies.
// This is explicit dependency injection — no magic, no container.
// The caller (main.go) creates the dependencies and passes them in.
func NewService(database *db.DB, repo *git.Repo, ollamaClient *ollama.Client, vaultDir string) *Service {
	return &Service{
		db:       database,
		git:      repo,
		ollama:   ollamaClient,
		vaultDir: vaultDir,
	}
}

// writeRawFile writes raw content to a note's file on disk.
// Used for frontmatter updates that bypass the normal write pipeline.
func (s *Service) writeRawFile(id, content string) error {
	note, err := s.db.GetNote(id)
	if err != nil || note == nil {
		return fmt.Errorf("note not found: %s", id)
	}

	fullPath := filepath.Join(s.vaultDir, note.FilePath)
	if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("write file: %w", err)
	}

	// Commit the update.
	return s.git.CommitFile(note.FilePath, fmt.Sprintf("vault: %s [import metadata]", note.Title))
}
