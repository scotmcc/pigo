// Package fetch retrieves web content and converts HTML to markdown.
// Layer-1 package — talks to HTTP and nothing else.
package fetch

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

// client is a shared HTTP client with a reasonable timeout.
var client = &http.Client{Timeout: 30 * time.Second}

// URL fetches the raw HTML content of a URL.
func URL(url string) (string, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	// Set a browser-like user agent to avoid being blocked.
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; pigo/1.0)")

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("fetch %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("fetch %s: status %d", url, resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read body: %w", err)
	}

	return string(body), nil
}

// Title extracts the <title> text from HTML.
// Returns empty string if no title found.
func Title(html string) string {
	return extractTag(html, "title")
}
