// Package releases queries external registries for the latest published
// version of pigo and its integrations. All functions use short timeouts
// and return errors that should be treated as non-fatal — an unreachable
// registry is an inconvenience, never a failure of the running binary.
package releases

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// githubAPI is the base URL for the GitHub REST API. Exposed as a var so
// tests can point it at a stub server.
var githubAPI = "https://api.github.com"

// PigoRelease is the subset of a GitHub release we care about.
type PigoRelease struct {
	TagName string // e.g. "v0.4.0"
	HTMLURL string // release page
}

// LatestPigo fetches the latest published pigo release from GitHub.
// Returns a 404-equivalent error if the repo has no releases yet.
func LatestPigo(ctx context.Context, owner, repo string) (PigoRelease, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/releases/latest", githubAPI, owner, repo)

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return PigoRelease{}, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return PigoRelease{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return PigoRelease{}, fmt.Errorf("no releases published yet for %s/%s", owner, repo)
	}
	if resp.StatusCode != http.StatusOK {
		return PigoRelease{}, fmt.Errorf("github returned %d", resp.StatusCode)
	}

	var body struct {
		TagName string `json:"tag_name"`
		HTMLURL string `json:"html_url"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return PigoRelease{}, fmt.Errorf("decode release: %w", err)
	}
	return PigoRelease{TagName: body.TagName, HTMLURL: body.HTMLURL}, nil
}
