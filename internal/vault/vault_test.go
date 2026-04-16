package vault

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/scotmcc/pigo/internal/db"
	"github.com/scotmcc/pigo/internal/git"
	"github.com/scotmcc/pigo/internal/ollama"
)

// setupTestVault creates a vault service backed by a temp directory.
// Ollama points at localhost — if it's not running, embeddings will be nil
// (which is fine, we test fuzzy search).
func setupTestVault(t *testing.T) (*Service, func()) {
	t.Helper()

	dir := t.TempDir()
	vaultDir := filepath.Join(dir, "vault")
	dbPath := filepath.Join(dir, "test.db")

	os.MkdirAll(vaultDir, 0755)

	database, err := db.Open(dbPath)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := database.Migrate(); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	repo, err := git.Open(vaultDir)
	if err != nil {
		t.Fatalf("open git: %v", err)
	}

	// Ollama client — will fail gracefully if not running.
	client := ollama.NewClient("http://localhost:11434", "nomic-embed-text")

	svc := NewService(database, repo, client, vaultDir)

	cleanup := func() {
		database.Close()
	}

	return svc, cleanup
}

func TestVaultWriteAndRead(t *testing.T) {
	svc, cleanup := setupTestVault(t)
	defer cleanup()

	// Write.
	result, err := svc.Write(WriteInput{
		Title: "Test Note",
		Tags:  []string{"test", "go"},
		Body:  "This is a test note.\n\n## Section One\n\nContent here.",
	})
	if err != nil {
		t.Fatalf("write: %v", err)
	}
	if result.ID != "test-note" {
		t.Errorf("id = %q, want %q", result.ID, "test-note")
	}

	// Read.
	note, err := svc.Read("test-note")
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if note.Title != "Test Note" {
		t.Errorf("title = %q", note.Title)
	}
	if len(note.Tags) != 2 {
		t.Errorf("tags = %v, want 2 tags", note.Tags)
	}

	// Verify file exists on disk.
	filePath := filepath.Join(svc.vaultDir, result.FilePath)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Error("file not created on disk")
	}
}

func TestVaultEdit(t *testing.T) {
	svc, cleanup := setupTestVault(t)
	defer cleanup()

	// Write initial note.
	svc.Write(WriteInput{Title: "Edit Me", Tags: []string{"v1"}, Body: "Original body."})

	// Edit.
	newBody := "Updated body."
	err := svc.Edit(EditInput{
		ID:   "edit-me",
		Body: &newBody,
		Tags: []string{"v2", "edited"},
	})
	if err != nil {
		t.Fatalf("edit: %v", err)
	}

	// Read back.
	note, err := svc.Read("edit-me")
	if err != nil {
		t.Fatalf("read after edit: %v", err)
	}
	if note.Tags[0] != "v2" {
		t.Errorf("tags not updated: %v", note.Tags)
	}
}

func TestVaultSearch(t *testing.T) {
	svc, cleanup := setupTestVault(t)
	defer cleanup()

	svc.Write(WriteInput{Title: "Go Patterns", Tags: []string{"go"}, Body: "Patterns in Go."})
	svc.Write(WriteInput{Title: "Rust Patterns", Tags: []string{"rust"}, Body: "Patterns in Rust."})

	// Fuzzy search by tag.
	result, err := svc.Search("go", 10)
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if len(result.Results) == 0 {
		t.Fatal("expected at least 1 result for 'go'")
	}
	if result.Results[0].NoteID != "go-patterns" {
		t.Errorf("top result = %q, want %q", result.Results[0].NoteID, "go-patterns")
	}
}

func TestVaultList(t *testing.T) {
	svc, cleanup := setupTestVault(t)
	defer cleanup()

	svc.Write(WriteInput{Title: "Note A", Body: "A"})
	svc.Write(WriteInput{Title: "Note B", Body: "B"})

	items, err := svc.List()
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(items) != 2 {
		t.Errorf("expected 2 notes, got %d", len(items))
	}
}

func TestVaultWriteDuplicate(t *testing.T) {
	svc, cleanup := setupTestVault(t)
	defer cleanup()

	svc.Write(WriteInput{Title: "Unique", Body: "First."})

	_, err := svc.Write(WriteInput{Title: "Unique", Body: "Second."})
	if err == nil {
		t.Error("expected error for duplicate title, got nil")
	}
}
