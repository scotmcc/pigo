package vault

import (
	"fmt"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// Frontmatter is the YAML metadata at the top of each note.
type Frontmatter struct {
	Title     string   `yaml:"title"`
	Tags      []string `yaml:"tags"`
	CreatedAt string   `yaml:"created"`
	UpdatedAt string   `yaml:"updated"`
	RelatesTo []string `yaml:"relates_to,omitempty"`
}

// ParseFrontmatter extracts YAML frontmatter from a markdown string.
// Returns the frontmatter and the remaining body.
func ParseFrontmatter(content string) (Frontmatter, string, error) {
	var fm Frontmatter

	if !strings.HasPrefix(content, "---\n") {
		return fm, content, nil
	}

	end := strings.Index(content[4:], "\n---")
	if end < 0 {
		return fm, content, nil
	}

	yamlBlock := content[4 : end+4]
	body := strings.TrimLeft(content[end+4+4:], "\n")

	if err := yaml.Unmarshal([]byte(yamlBlock), &fm); err != nil {
		return fm, content, fmt.Errorf("parse frontmatter: %w", err)
	}

	return fm, body, nil
}

// RenderNote produces the full markdown file content: frontmatter + body.
func RenderNote(fm Frontmatter, body string) (string, error) {
	yamlBytes, err := yaml.Marshal(fm)
	if err != nil {
		return "", fmt.Errorf("marshal frontmatter: %w", err)
	}

	var b strings.Builder
	b.WriteString("---\n")
	b.Write(yamlBytes)
	b.WriteString("---\n\n")
	b.WriteString(body)

	if !strings.HasSuffix(body, "\n") {
		b.WriteString("\n")
	}

	return b.String(), nil
}

// NewFrontmatter creates frontmatter with timestamps set to now.
func NewFrontmatter(title string, tags []string) Frontmatter {
	now := time.Now().UTC().Format(time.RFC3339)
	return Frontmatter{
		Title:     title,
		Tags:      tags,
		CreatedAt: now,
		UpdatedAt: now,
	}
}
