// Package vault implements the vault operations: read, write, edit, search, import, links.
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
type Service struct {
	db       *db.DB
	git      *git.Repo
	ollama   *ollama.Client
	vaultDir string
}

// NewService creates a vault service with its dependencies.
func NewService(database *db.DB, repo *git.Repo, ollamaClient *ollama.Client, vaultDir string) *Service {
	return &Service{
		db:       database,
		git:      repo,
		ollama:   ollamaClient,
		vaultDir: vaultDir,
	}
}

// updateNoteFile re-renders a note's frontmatter + body and commits the change.
// Used for post-write updates like relationship discovery.
func (s *Service) updateNoteFile(id string, fm Frontmatter, body string, commitMsg string) error {
	note, err := s.db.GetNote(id)
	if err != nil || note == nil {
		return fmt.Errorf("note not found: %s", id)
	}

	content, err := RenderNote(fm, body)
	if err != nil {
		return err
	}

	fullPath := filepath.Join(s.vaultDir, note.FilePath)
	if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("write file: %w", err)
	}

	return s.git.CommitFile(note.FilePath, commitMsg)
}
