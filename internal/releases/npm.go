package releases

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os/exec"
	"time"
)

// npmRegistry is the base URL for the npm registry. Exposed as a var so tests
// can point it at a stub server.
var npmRegistry = "https://registry.npmjs.org"

// LatestNpm returns the "latest" dist-tag for the given npm package (which can
// include an @scope/name). Uses the lightweight dist-tags endpoint so we
// don't download full package metadata.
func LatestNpm(ctx context.Context, pkg string) (string, error) {
	endpoint := fmt.Sprintf("%s/-/package/%s/dist-tags", npmRegistry, url.PathEscape(pkg))

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return "", err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("npm registry returned %d", resp.StatusCode)
	}

	var body map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return "", fmt.Errorf("decode dist-tags: %w", err)
	}
	latest, ok := body["latest"]
	if !ok || latest == "" {
		return "", fmt.Errorf("no 'latest' dist-tag for %s", pkg)
	}
	return latest, nil
}

// InstalledNpm runs `npm ls -g <pkg> --json --depth=0` and returns the version
// of the globally-installed package. Returns a non-nil error if the package
// is not installed or npm itself is missing.
func InstalledNpm(pkg string) (string, error) {
	out, err := exec.Command("npm", "ls", "-g", pkg, "--json", "--depth=0").Output()
	if err != nil {
		// npm ls returns non-zero when the package isn't installed, but it
		// still writes valid JSON. Try to decode before giving up on err.
		if ee, ok := err.(*exec.ExitError); ok {
			out = append(out, ee.Stderr...)
		} else if len(out) == 0 {
			return "", fmt.Errorf("run npm ls: %w", err)
		}
	}

	var body struct {
		Dependencies map[string]struct {
			Version string `json:"version"`
		} `json:"dependencies"`
	}
	if err := json.Unmarshal(out, &body); err != nil {
		return "", fmt.Errorf("decode npm ls output: %w", err)
	}

	entry, ok := body.Dependencies[pkg]
	if !ok || entry.Version == "" {
		return "", fmt.Errorf("%s not installed globally", pkg)
	}
	return entry.Version, nil
}
