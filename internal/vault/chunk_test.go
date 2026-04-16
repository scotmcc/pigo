package vault

import "testing"

func TestChunkMarkdown_NoHeadings(t *testing.T) {
	chunks := ChunkMarkdown("Just a paragraph of text.")
	if len(chunks) != 1 {
		t.Fatalf("expected 1 chunk, got %d", len(chunks))
	}
	if chunks[0].Heading != "" {
		t.Errorf("expected empty heading, got %q", chunks[0].Heading)
	}
	if chunks[0].Content != "Just a paragraph of text." {
		t.Errorf("unexpected content: %q", chunks[0].Content)
	}
}

func TestChunkMarkdown_WithHeadings(t *testing.T) {
	md := `Intro paragraph.

## First Section

Content of first section.

## Second Section

Content of second section.`

	chunks := ChunkMarkdown(md)
	if len(chunks) != 3 {
		t.Fatalf("expected 3 chunks (intro + 2 sections), got %d", len(chunks))
	}

	// Intro chunk.
	if chunks[0].Heading != "" {
		t.Errorf("intro chunk should have empty heading, got %q", chunks[0].Heading)
	}
	if chunks[0].Content != "Intro paragraph." {
		t.Errorf("intro content: %q", chunks[0].Content)
	}

	// First section.
	if chunks[1].Heading != "First Section" {
		t.Errorf("heading = %q, want %q", chunks[1].Heading, "First Section")
	}
	if chunks[1].Anchor != "#first-section" {
		t.Errorf("anchor = %q, want %q", chunks[1].Anchor, "#first-section")
	}

	// Second section.
	if chunks[2].Heading != "Second Section" {
		t.Errorf("heading = %q, want %q", chunks[2].Heading, "Second Section")
	}
}

func TestChunkMarkdown_Empty(t *testing.T) {
	chunks := ChunkMarkdown("")
	if len(chunks) != 0 {
		t.Fatalf("expected 0 chunks for empty input, got %d", len(chunks))
	}
}

func TestChunkMarkdown_HeadingOnly(t *testing.T) {
	chunks := ChunkMarkdown("## Just a Heading")
	if len(chunks) != 1 {
		t.Fatalf("expected 1 chunk, got %d", len(chunks))
	}
	if chunks[0].Heading != "Just a Heading" {
		t.Errorf("heading = %q", chunks[0].Heading)
	}
}
