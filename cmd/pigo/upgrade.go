package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/scotmcc/pigo/internal/detect"
	"github.com/scotmcc/pigo/internal/keys"
	"github.com/scotmcc/pigo/internal/releases"
	"github.com/spf13/cobra"
	"golang.org/x/mod/semver"
)

// pigoRepoOwner and pigoRepoName point at the canonical release source.
// If the project moves, update these constants in one place.
const (
	pigoRepoOwner = "scotmcc"
	pigoRepoName  = "pigo"
)

var upgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "Check for a newer pigo and pi, then refresh integrations",
	Long: `Checks GitHub for a newer pigo release and the npm registry for a newer pi.
Offers to upgrade pi automatically (it's managed by npm). pigo itself must be
upgraded manually — the running binary can't safely overwrite itself.

After the version checks, runs 'pigo install' to refresh the pi extension and
Claude Code skill from the currently-installed binary.`,
	RunE: runUpgrade,
}

func init() {
	rootCmd.AddCommand(upgradeCmd)
}

func runUpgrade(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	checkPigoVersion(ctx)
	offerPiUpgrade(ctx)

	fmt.Println()
	return runInstall(cmd, args)
}

// checkPigoVersion prints a banner if a newer pigo release is available.
// Never fails the command — network issues are inconvenient, not fatal.
func checkPigoVersion(ctx context.Context) {
	rel, err := releases.LatestPigo(ctx, pigoRepoOwner, pigoRepoName)
	if err != nil {
		fmt.Printf("couldn't check for a newer pigo: %v\n", err)
		return
	}

	current := toSemver(keys.Version)
	latest := toSemver(rel.TagName)

	// If either side isn't valid semver we can't compare. Show the fact, not noise.
	if !semver.IsValid(current) || !semver.IsValid(latest) {
		fmt.Printf("running pigo %s — latest release is %s (skipping version compare)\n", keys.Version, rel.TagName)
		return
	}

	switch semver.Compare(current, latest) {
	case -1:
		fmt.Println("Update Available")
		fmt.Printf("  pigo %s is available (you have %s)\n", rel.TagName, keys.Version)
		fmt.Printf("  Changelog: %s\n", rel.HTMLURL)
		fmt.Printf("  Install:   go install github.com/%s/%s/cmd/pigo@latest\n", pigoRepoOwner, pigoRepoName)
		fmt.Printf("             (or download: https://github.com/%s/%s/releases/latest)\n", pigoRepoOwner, pigoRepoName)
	case 0:
		fmt.Printf("pigo %s is the latest.\n", keys.Version)
	case 1:
		fmt.Printf("running pigo %s (ahead of latest release %s — dev build)\n", keys.Version, rel.TagName)
	}
}

// offerPiUpgrade checks the npm registry, compares to the globally-installed
// pi version, and prompts the user to upgrade if a newer one is available.
// Silently skips when pi or npm isn't present — Chunk D's install flow owns
// the "pi missing" experience.
func offerPiUpgrade(ctx context.Context) {
	if !detect.OnPath("pi") || !detect.OnPath("npm") {
		return
	}

	installed, err := releases.InstalledNpm(piNpmPkg)
	if err != nil {
		// pi is on PATH but wasn't installed via global npm — nothing we can do.
		return
	}
	latest, err := releases.LatestNpm(ctx, piNpmPkg)
	if err != nil {
		fmt.Printf("couldn't check for a newer pi: %v\n", err)
		return
	}

	if semver.Compare(toSemver(installed), toSemver(latest)) >= 0 {
		fmt.Printf("pi %s is the latest.\n", installed)
		return
	}

	fmt.Println()
	fmt.Printf("pi %s is available (you have %s).\n", latest, installed)
	cmd := fmt.Sprintf("npm install -g %s@latest", piNpmPkg)
	if !confirm("Upgrade pi now?", true) {
		fmt.Printf("skipped pi upgrade — run '%s' yourself when ready\n", cmd)
		return
	}

	fmt.Printf("\nrunning: %s\n\n", cmd)
	if err := runShell(cmd); err != nil {
		fmt.Printf("pi upgrade failed: %v — run '%s' manually\n", err, cmd)
		return
	}
	fmt.Printf("upgraded pi to %s\n", latest)
}

// toSemver normalizes a version string so golang.org/x/mod/semver can parse it.
// Adds a leading 'v' if missing (npm versions come as "0.67.1", git tags as "v0.67.1").
func toSemver(v string) string {
	if !strings.HasPrefix(v, "v") {
		return "v" + v
	}
	return v
}
