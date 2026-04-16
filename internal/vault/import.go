package vault

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/scotmcc/pigo/internal/fetch"
)

// ImportInput is what the caller provides to import a URL as a note.
type ImportInput struct {
	URL  string
	Tags []string // additional tags (domain is auto-added)
}

// Import fetches a URL, converts the HTML to markdown, and saves it as a vault note.
// The note gets the page title, auto-tags with the domain and "imported", and
// stores the source URL in frontmatter.
func (s *Service) Import(input ImportInput) (*WriteResult, error) {
	// Fetch the page.
	html, err := fetch.URL(input.URL)
	if err != nil {
		return nil, fmt.Errorf("fetch: %w", err)
	}

	// Extract title from HTML.
	title := fetch.Title(html)
	if title == "" {
		// Fall back to the URL path.
		title = titleFromURL(input.URL)
	}

	// Convert to markdown.
	body := fetch.ToMarkdown(html)
	if body == "" {
		return nil, fmt.Errorf("no content extracted from %s", input.URL)
	}

	// Build tags: user-provided + domain + "imported".
	tags := append([]string{}, input.Tags...)
	if domain := domainFromURL(input.URL); domain != "" {
		tags = append(tags, domain)
	}
	tags = append(tags, "imported")

	// Write as a vault note.
	result, err := s.Write(WriteInput{
		Title: title,
		Tags:  tags,
		Body:  body,
	})
	if err != nil {
		return nil, err
	}

	// Update frontmatter with source_url.
	// We read back, add the field, and re-save.
	if err := s.addSourceURL(result.ID, input.URL); err != nil {
		// Non-fatal — the note is saved, just missing source_url.
		fmt.Printf("warning: could not add source_url to frontmatter: %v\n", err)
	}

	return result, nil
}

// addSourceURL reads a note's file, adds source_url to frontmatter, and re-saves.
func (s *Service) addSourceURL(id, sourceURL string) error {
	note, err := s.Read(id)
	if err != nil {
		return err
	}

	fm, body, err := ParseFrontmatter(note.RawContent)
	if err != nil {
		return err
	}

	// Add source_url as a custom field by re-rendering with it in the YAML.
	// Since our Frontmatter struct doesn't have SourceURL, we append it manually.
	rendered, err := RenderNote(fm, body)
	if err != nil {
		return err
	}

	// Insert source_url after the last frontmatter field.
	rendered = strings.Replace(rendered, "---\n\n", fmt.Sprintf("source_url: %s\n---\n\n", sourceURL), 1)

	return s.writeRawFile(id, rendered)
}

func domainFromURL(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}
	host := u.Hostname()
	// Remove www. prefix.
	host = strings.TrimPrefix(host, "www.")
	return host
}

func titleFromURL(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}
	path := strings.Trim(u.Path, "/")
	if path == "" {
		return u.Hostname()
	}
	// Use the last path segment, replace hyphens with spaces.
	parts := strings.Split(path, "/")
	last := parts[len(parts)-1]
	last = strings.ReplaceAll(last, "-", " ")
	last = strings.ReplaceAll(last, "_", " ")
	// Remove file extension.
	if idx := strings.LastIndex(last, "."); idx > 0 {
		last = last[:idx]
	}
	return last
}
