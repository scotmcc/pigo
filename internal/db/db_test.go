package db

import (
	"path/filepath"
	"testing"
)

// TestSQLiteVecLoaded verifies that sqlite-vec is bundled and loaded on every
// connection — semantic search depends on vec_distance_cosine being available.
func TestSQLiteVecLoaded(t *testing.T) {
	d, err := Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer d.Close()

	var vecVersion string
	if err := d.Conn().QueryRow("SELECT vec_version()").Scan(&vecVersion); err != nil {
		t.Fatalf("vec_version() failed — sqlite-vec not loaded: %v", err)
	}
	if vecVersion == "" {
		t.Fatal("vec_version() returned empty string")
	}
}

// TestVecDistanceCosine confirms the cosine distance function used by VectorSearch
// is callable end to end. Uses two identical 3-dim vectors (distance should be 0).
func TestVecDistanceCosine(t *testing.T) {
	d, err := Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer d.Close()

	var distance float64
	err = d.Conn().QueryRow(
		"SELECT vec_distance_cosine(vec_f32(?), vec_f32(?))",
		"[1.0, 2.0, 3.0]", "[1.0, 2.0, 3.0]",
	).Scan(&distance)
	if err != nil {
		t.Fatalf("vec_distance_cosine failed: %v", err)
	}
	if distance > 0.0001 {
		t.Errorf("identical vectors should have distance ~0, got %v", distance)
	}
}
