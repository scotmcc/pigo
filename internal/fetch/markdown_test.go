package fetch

import (
	"strings"
	"testing"
)

func TestToMarkdown_Headings(t *testing.T) {
	html := `<h1>Title</h1><h2>Subtitle</h2><h3>Section</h3>`
	md := ToMarkdown(html)

	if !strings.Contains(md, "# Title") {
		t.Error("expected # Title")
	}
	if !strings.Contains(md, "## Subtitle") {
		t.Error("expected ## Subtitle")
	}
	if !strings.Contains(md, "### Section") {
		t.Error("expected ### Section")
	}
}

func TestToMarkdown_Paragraphs(t *testing.T) {
	html := `<p>First paragraph.</p><p>Second paragraph.</p>`
	md := ToMarkdown(html)

	if !strings.Contains(md, "First paragraph.") {
		t.Error("missing first paragraph")
	}
	if !strings.Contains(md, "Second paragraph.") {
		t.Error("missing second paragraph")
	}
}

func TestToMarkdown_Links(t *testing.T) {
	html := `<a href="https://go.dev">Go website</a>`
	md := ToMarkdown(html)

	if !strings.Contains(md, "[Go website](https://go.dev)") {
		t.Errorf("link not converted: %q", md)
	}
}

func TestToMarkdown_Formatting(t *testing.T) {
	html := `<strong>bold</strong> and <em>italic</em> and <code>code</code>`
	md := ToMarkdown(html)

	if !strings.Contains(md, "**bold**") {
		t.Error("bold not converted")
	}
	if !strings.Contains(md, "*italic*") {
		t.Error("italic not converted")
	}
	if !strings.Contains(md, "`code`") {
		t.Error("code not converted")
	}
}

func TestToMarkdown_Lists(t *testing.T) {
	html := `<ul><li>One</li><li>Two</li><li>Three</li></ul>`
	md := ToMarkdown(html)

	if strings.Count(md, "- ") != 3 {
		t.Errorf("expected 3 list items, got: %q", md)
	}
}

func TestToMarkdown_StripScripts(t *testing.T) {
	html := `<p>Hello</p><script>alert('xss')</script><p>World</p>`
	md := ToMarkdown(html)

	if strings.Contains(md, "alert") {
		t.Error("script content not stripped")
	}
	if !strings.Contains(md, "Hello") || !strings.Contains(md, "World") {
		t.Error("content around script lost")
	}
}

func TestToMarkdown_Entities(t *testing.T) {
	html := `<p>A &amp; B &lt; C &gt; D</p>`
	md := ToMarkdown(html)

	if !strings.Contains(md, "A & B < C > D") {
		t.Errorf("entities not decoded: %q", md)
	}
}

func TestTitle(t *testing.T) {
	html := `<html><head><title>My Page Title</title></head><body></body></html>`
	title := Title(html)

	if title != "My Page Title" {
		t.Errorf("title = %q, want %q", title, "My Page Title")
	}
}

func TestTitle_Missing(t *testing.T) {
	html := `<html><body>No title here</body></html>`
	title := Title(html)

	if title != "" {
		t.Errorf("expected empty title, got %q", title)
	}
}
