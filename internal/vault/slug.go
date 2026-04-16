package vault

import (
	"regexp"
	"strings"
)

// nonAlphaNum matches anything that isn't a letter, number, or hyphen.
var nonAlphaNum = regexp.MustCompile(`[^a-z0-9-]+`)

// multiHyphen matches consecutive hyphens.
var multiHyphen = regexp.MustCompile(`-{2,}`)

// slugify converts a string to a URL/filesystem-safe slug.
// "Why We Chose Go" → "why-we-chose-go"
func slugify(s string) string {
	s = strings.ToLower(s)
	s = nonAlphaNum.ReplaceAllString(s, "-")
	s = multiHyphen.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	return s
}
