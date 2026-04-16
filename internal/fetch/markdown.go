package fetch

import (
	"fmt"
	"regexp"
	"strings"
)

// ToMarkdown converts HTML to clean markdown using regex replacements.
// Not a full DOM parser — good enough for knowledge capture from articles.
func ToMarkdown(html string) string {
	s := html

	// 1. Remove noise: scripts, styles, nav, footer, header.
	for _, tag := range []string{"script", "style", "nav", "footer", "header", "noscript"} {
		s = removeTag(s, tag)
	}

	// 2. Convert headings (h1-h6 → # prefixes).
	for i := 1; i <= 6; i++ {
		prefix := strings.Repeat("#", i)
		s = replaceTag(s, fmt.Sprintf("h%d", i), prefix+" $1\n\n")
	}

	// 3. Convert common elements.
	s = replaceTag(s, "p", "$1\n\n")
	s = replaceTag(s, "li", "- $1\n")
	s = replaceTag(s, "strong", "**$1**")
	s = replaceTag(s, "b", "**$1**")
	s = replaceTag(s, "em", "*$1*")
	s = replaceTag(s, "i", "*$1*")
	s = replaceTag(s, "code", "`$1`")
	s = replaceTagRe(s, `(?i)<pre[^>]*>(.*?)</pre>`, "```\n$1\n```\n\n")

	// 4. Convert links: <a href="url">text</a> → [text](url)
	s = linkRe.ReplaceAllString(s, "[$2]($1)")

	// 5. Convert line breaks.
	s = brRe.ReplaceAllString(s, "\n")
	s = divCloseRe.ReplaceAllString(s, "\n")

	// 6. Strip remaining HTML tags.
	s = tagRe.ReplaceAllString(s, "")

	// 7. Decode HTML entities.
	s = decodeEntities(s)

	// 8. Clean up whitespace.
	s = multiNewline.ReplaceAllString(s, "\n\n")
	s = trailingSpaces.ReplaceAllString(s, "")
	s = strings.TrimSpace(s)

	return s
}

// --- regex patterns (compiled once) ---

var (
	linkRe         = regexp.MustCompile(`(?is)<a\s+[^>]*href=["']([^"']+)["'][^>]*>(.*?)</a>`)
	brRe           = regexp.MustCompile(`(?i)<br\s*/?>`)
	divCloseRe     = regexp.MustCompile(`(?i)</div>`)
	tagRe          = regexp.MustCompile(`<[^>]+>`)
	multiNewline   = regexp.MustCompile(`\n{3,}`)
	trailingSpaces = regexp.MustCompile(`(?m)[ \t]+$`)
)

// removeTag removes a tag and everything between its open and close.
func removeTag(s, tag string) string {
	re := regexp.MustCompile(fmt.Sprintf(`(?is)<%s[^>]*>.*?</%s>`, tag, tag))
	return re.ReplaceAllString(s, "")
}

// replaceTag converts <tag>content</tag> to a replacement pattern.
// $1 in the replacement refers to the inner content.
func replaceTag(s, tag, replacement string) string {
	re := regexp.MustCompile(fmt.Sprintf(`(?is)<%s[^>]*>(.*?)</%s>`, tag, tag))
	return re.ReplaceAllString(s, replacement)
}

// replaceTagRe applies a custom regex pattern and replacement.
func replaceTagRe(s, pattern, replacement string) string {
	re := regexp.MustCompile(pattern)
	return re.ReplaceAllString(s, replacement)
}

// decodeEntities converts common HTML entities to their characters.
func decodeEntities(s string) string {
	replacer := strings.NewReplacer(
		"&nbsp;", " ",
		"&lt;", "<",
		"&gt;", ">",
		"&amp;", "&",
		"&quot;", `"`,
		"&#39;", "'",
		"&#x27;", "'",
		"&mdash;", "—",
		"&ndash;", "–",
		"&hellip;", "...",
		"&copy;", "(c)",
		"&reg;", "(R)",
	)
	return replacer.Replace(s)
}

// extractTag returns the inner text of the first occurrence of a tag.
func extractTag(html, tag string) string {
	re := regexp.MustCompile(fmt.Sprintf(`(?is)<%s[^>]*>(.*?)</%s>`, tag, tag))
	match := re.FindStringSubmatch(html)
	if len(match) < 2 {
		return ""
	}
	// Strip any nested tags from the result.
	return strings.TrimSpace(tagRe.ReplaceAllString(match[1], ""))
}

