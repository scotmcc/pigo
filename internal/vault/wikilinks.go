package vault

import "regexp"

// wikiLinkRe matches [[slug]] or [[slug|display text]] in note bodies.
// This is the Obsidian-style linking syntax.
var wikiLinkRe = regexp.MustCompile(`\[\[([^\]|]+)(?:\|[^\]]+)?\]\]`)

// DetectWikiLinks finds all [[slug]] references in a body and returns
// the unique slugs. These become the links_to field in frontmatter.
func DetectWikiLinks(body string) []string {
	matches := wikiLinkRe.FindAllStringSubmatch(body, -1)
	if len(matches) == 0 {
		return nil
	}

	// Deduplicate.
	seen := make(map[string]bool)
	var links []string
	for _, m := range matches {
		slug := slugify(m[1])
		if slug != "" && !seen[slug] {
			seen[slug] = true
			links = append(links, slug)
		}
	}

	return links
}
