package vault

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/scotmcc/pigo/internal/fetch"
)

// ImportInput is what the caller provides to import a URL as a note.
type ImportInput struct {
	URL   string
	Tags  []string // additional tags (domain is auto-added)
	Query string   // the search query that led to this import (optional)
}

// Import fetches a URL, converts the HTML to markdown, and saves it as a vault note.
func (s *Service) Import(input ImportInput) (*WriteResult, error) {
	html, err := fetch.URL(input.URL)
	if err != nil {
		return nil, fmt.Errorf("fetch: %w", err)
	}

	title := fetch.Title(html)
	if title == "" {
		title = titleFromURL(input.URL)
	}

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

	// Use the import-specific frontmatter (includes type + source_url).
	return s.WriteWithFrontmatter(
		NewImportFrontmatter(title, tags, input.URL),
		body,
	)
}

func domainFromURL(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}
	return strings.TrimPrefix(u.Hostname(), "www.")
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
	parts := strings.Split(path, "/")
	last := parts[len(parts)-1]
	last = strings.ReplaceAll(last, "-", " ")
	last = strings.ReplaceAll(last, "_", " ")
	if idx := strings.LastIndex(last, "."); idx > 0 {
		last = last[:idx]
	}
	return last
}
