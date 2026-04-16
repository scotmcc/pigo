package vault

import (
	"strings"
	"testing"
)

func TestParseFrontmatter(t *testing.T) {
	input := `---
title: Test Note
tags:
    - go
    - testing
created: "2026-04-16T10:00:00Z"
updated: "2026-04-16T10:00:00Z"
---

This is the body.`

	fm, body, err := ParseFrontmatter(input)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	if fm.Title != "Test Note" {
		t.Errorf("title = %q, want %q", fm.Title, "Test Note")
	}
	if len(fm.Tags) != 2 || fm.Tags[0] != "go" || fm.Tags[1] != "testing" {
		t.Errorf("tags = %v, want [go testing]", fm.Tags)
	}
	if !strings.Contains(body, "This is the body.") {
		t.Errorf("body = %q, expected to contain 'This is the body.'", body)
	}
}

func TestParseFrontmatter_NoFrontmatter(t *testing.T) {
	input := "Just plain markdown."
	fm, body, err := ParseFrontmatter(input)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	if fm.Title != "" {
		t.Errorf("expected empty title, got %q", fm.Title)
	}
	if body != input {
		t.Errorf("body should equal input when no frontmatter")
	}
}

func TestRenderNote(t *testing.T) {
	fm := NewFrontmatter("Test", []string{"a", "b"})
	content, err := RenderNote(fm, "Hello world")
	if err != nil {
		t.Fatalf("render error: %v", err)
	}

	if !strings.HasPrefix(content, "---\n") {
		t.Error("should start with ---")
	}
	if !strings.Contains(content, "title: Test") {
		t.Error("should contain title")
	}
	if !strings.Contains(content, "Hello world") {
		t.Error("should contain body")
	}
}

func TestRoundTrip(t *testing.T) {
	fm := NewFrontmatter("Round Trip", []string{"test"})
	original := "Some content here."

	rendered, err := RenderNote(fm, original)
	if err != nil {
		t.Fatalf("render: %v", err)
	}

	parsed, body, err := ParseFrontmatter(rendered)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	if parsed.Title != "Round Trip" {
		t.Errorf("title = %q after round trip", parsed.Title)
	}
	if !strings.Contains(body, "Some content here.") {
		t.Errorf("body lost after round trip: %q", body)
	}
}
