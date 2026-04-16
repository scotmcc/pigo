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

// Write creates a new note in the vault using default frontmatter.
func (s *Service) Write(input WriteInput) (*WriteResult, error) {
	fm := NewFrontmatter(input.Title, input.Tags)
	return s.WriteWithFrontmatter(fm, input.Body)
}

// WriteWithFrontmatter creates a new note with custom frontmatter.
// Used by Import (which sets type, source_url, etc.) and anything else
// that needs non-default metadata.
func (s *Service) WriteWithFrontmatter(fm Frontmatter, body string) (*WriteResult, error) {
	slug := slugify(fm.Title)
	if slug == "" {
		return nil, fmt.Errorf("title produces empty slug")
	}

	existing, err := s.db.GetNote(slug)
	if err != nil {
		return nil, fmt.Errorf("check existing: %w", err)
	}
	if existing != nil {
		return nil, fmt.Errorf("note already exists: %s", slug)
	}

	// Detect [[wiki-links]] in the body and record them.
	fm.LinksTo = DetectWikiLinks(body)

	// Render markdown file.
	content, err := RenderNote(fm, body)
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
	tagsJSON, _ := json.Marshal(fm.Tags)
	now := time.Now().UTC()
	note := db.Note{
		ID:        slug,
		FilePath:  relPath,
		Title:     fm.Title,
		Tags:      string(tagsJSON),
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := s.db.InsertNote(note); err != nil {
		return nil, err
	}

	// Chunk, embed, and index.
	if err := s.indexChunks(slug, body); err != nil {
		return nil, fmt.Errorf("index chunks: %w", err)
	}

	// Git commit.
	if err := s.git.CommitFile(relPath, fmt.Sprintf("vault: %s [write]", fm.Title)); err != nil {
		return nil, fmt.Errorf("git commit: %w", err)
	}

	// Discover relationships in the background.
	// Non-fatal — the note is saved even if relationship discovery fails.
	if err := s.discoverRelationships(slug); err != nil {
		fmt.Fprintf(os.Stderr, "warning: relationship discovery failed: %v\n", err)
	}

	return &WriteResult{ID: slug, FilePath: relPath}, nil
}

// indexChunks splits the body, embeds each chunk, and stores them in SQLite.
func (s *Service) indexChunks(noteID string, body string) error {
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
