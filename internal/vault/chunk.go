package vault

import (
	"regexp"
	"strings"
)

// ChunkResult holds a parsed chunk from splitting a markdown file on headings.
type ChunkResult struct {
	Heading string // heading text, empty for intro chunk
	Anchor  string // #heading-slug for deep linking
	Content string // the text under this heading
}

// headingRegex matches markdown headings (## Heading Text).
var headingRegex = regexp.MustCompile(`(?m)^(#{1,6})\s+(.+)$`)

// ChunkMarkdown splits markdown content on headings.
// Text before the first heading becomes the "intro" chunk.
// Each heading starts a new chunk that includes all text until the next heading.
func ChunkMarkdown(body string) []ChunkResult {
	matches := headingRegex.FindAllStringIndex(body, -1)

	if len(matches) == 0 {
		// No headings — the whole body is one chunk.
		if trimmed := strings.TrimSpace(body); trimmed != "" {
			return []ChunkResult{{Content: trimmed}}
		}
		return nil
	}

	var chunks []ChunkResult

	// Intro chunk: text before the first heading.
	if matches[0][0] > 0 {
		intro := strings.TrimSpace(body[:matches[0][0]])
		if intro != "" {
			chunks = append(chunks, ChunkResult{Content: intro})
		}
	}

	// Each heading starts a chunk.
	for i, match := range matches {
		line := body[match[0]:match[1]]
		heading := headingRegex.FindStringSubmatch(line)
		if len(heading) < 3 {
			continue
		}

		title := heading[2]
		anchor := slugify(title)

		// Content runs from after the heading line to the next heading (or end).
		contentStart := match[1]
		contentEnd := len(body)
		if i+1 < len(matches) {
			contentEnd = matches[i+1][0]
		}
		content := strings.TrimSpace(body[contentStart:contentEnd])

		chunks = append(chunks, ChunkResult{
			Heading: title,
			Anchor:  "#" + anchor,
			Content: content,
		})
	}

	return chunks
}
