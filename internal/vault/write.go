package vault

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/scotmcc/pigo/internal/db"
)

// WriteInput is what the caller provides to create a new note.
type WriteInput struct {
	Title string
	Tags  []string
	Body  string
}

// WriteResult is returned after a successful write.
type WriteResult struct {
	ID       string
	FilePath string
}

// Write creates a new note in the vault.
// It saves the file, indexes it, embeds chunks, and commits to git.
func (s *Service) Write(input WriteInput) (*WriteResult, error) {
	slug := slugify(input.Title)
	if slug == "" {
		return nil, fmt.Errorf("title produces empty slug")
	}

	// Check for ID collision.
	existing, err := s.db.GetNote(slug)
	if err != nil {
		return nil, fmt.Errorf("check existing: %w", err)
	}
	if existing != nil {
		return nil, fmt.Errorf("note already exists: %s", slug)
	}

	// Build and render the markdown file.
	fm := NewFrontmatter(input.Title, input.Tags)
	content, err := RenderNote(fm, input.Body)
	if err != nil {
		return nil, err
	}

	// Write file to disk.
	relPath := slug + ".md"
	fullPath := filepath.Join(s.vaultDir, relPath)
	if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
		return nil, fmt.Errorf("write file: %w", err)
	}

	// Index in SQLite.
	tagsJSON, _ := json.Marshal(input.Tags)
	now := time.Now().UTC()
	note := db.Note{
		ID:        slug,
		FilePath:  relPath,
		Title:     input.Title,
		Tags:      string(tagsJSON),
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := s.db.InsertNote(note); err != nil {
		return nil, err
	}

	// Chunk, embed, and index.
	if err := s.indexChunks(slug, input.Body); err != nil {
		return nil, fmt.Errorf("index chunks: %w", err)
	}

	// Git commit.
	if err := s.git.CommitFile(relPath, fmt.Sprintf("vault: %s [write]", input.Title)); err != nil {
		return nil, fmt.Errorf("git commit: %w", err)
	}

	return &WriteResult{ID: slug, FilePath: relPath}, nil
}

// indexChunks splits the body, embeds each chunk, and stores them in SQLite.
func (s *Service) indexChunks(noteID string, body string) error {
	// Remove old chunks first (for re-indexing on edit).
	if err := s.db.DeleteChunksByNoteID(noteID); err != nil {
		return err
	}

	chunks := ChunkMarkdown(body)
	if len(chunks) == 0 {
		return nil
	}

	var dbChunks []db.Chunk
	for _, c := range chunks {
		embedding, err := s.ollama.Embed(c.Content)
		if err != nil {
			// Don't fail — the note is still useful without embeddings.
			// First failure logs the warning; subsequent chunks skip silently.
			if len(dbChunks) == 0 {
				fmt.Fprintf(os.Stderr, "warning: embeddings unavailable (Ollama not running) — note saved without semantic indexing\n")
			}
			embedding = nil
		}

		dbChunks = append(dbChunks, db.Chunk{
			ID:        uuid.New().String(),
			NoteID:    noteID,
			Heading:   c.Heading,
			Anchor:    c.Anchor,
			Content:   c.Content,
			Embedding: embedding,
		})
	}

	return s.db.InsertChunks(dbChunks)
}
