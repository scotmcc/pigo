package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/scotmcc/pigo/internal/assets"
	"github.com/scotmcc/pigo/internal/detect"
)

// piNpmPkg is the npm package name for the pi coding agent.
const piNpmPkg = "@mariozechner/pi-coding-agent"

// piStep detects pi and either installs pigo's pi extensions, offers to
// install pi itself via npm, or tells the user exactly what to run so they
// can get there themselves. Returns summary lines for the install report.
func piStep(home string) []string {
	piExtDir := filepath.Join(home, ".pi", "extensions")

	// pi is "present" if either the extensions dir already exists or the
	// binary is on PATH (the dir may not exist yet on a fresh pi install).
	if dirExists(piExtDir) || detect.OnPath("pi") {
		return installPiExtensions(piExtDir)
	}

	return offerPiInstall(piExtDir)
}

// installPiExtensions drops pigo's embedded pi extensions into ~/.pi/extensions/,
// creating the directory first if needed.
func installPiExtensions(piExtDir string) []string {
	if err := os.MkdirAll(piExtDir, 0755); err != nil {
		return []string{fmt.Sprintf("failed to create %s: %v", piExtDir, err)}
	}

	var summary []string

	piDest := filepath.Join(piExtDir, "pigo.ts")
	if err := os.WriteFile(piDest, []byte(assets.PiExtension), 0644); err != nil {
		return []string{fmt.Sprintf("write pi extension: %v", err)}
	}
	summary = append(summary, fmt.Sprintf("installed pi extension → %s", piDest))

	ollamaDest := filepath.Join(piExtDir, "ollama.js")
	if err := os.WriteFile(ollamaDest, []byte(assets.OllamaExtension), 0644); err != nil {
		return append(summary, fmt.Sprintf("write ollama extension: %v", err))
	}
	summary = append(summary, fmt.Sprintf("installed ollama provider → %s", ollamaDest))

	return summary
}

// offerPiInstall is the "pi absent" branch. If npm is available we offer to
// npm-install pi globally; otherwise we point at node/nvm with the exact
// command to run once node is on PATH.
func offerPiInstall(piExtDir string) []string {
	cmd := "npm install -g " + piNpmPkg

	if !detect.OnPath("npm") {
		hint := "install Node.js (via nvm or your package manager)"
		if detect.IsMac() {
			hint = "install Node.js (e.g., 'brew install node' or via nvm)"
		}
		return []string{fmt.Sprintf("pi not installed and npm missing — %s, then run '%s', then re-run 'pigo install'", hint, cmd)}
	}

	fmt.Println()
	fmt.Printf("pi coding agent is not installed. pigo can install it via npm:\n  %s\n", cmd)
	if !confirm("Run this now?", false) {
		return []string{fmt.Sprintf("skipped pi install — run '%s' yourself, then re-run 'pigo install'", cmd)}
	}

	fmt.Printf("\nrunning: %s\n\n", cmd)
	if err := runShell(cmd); err != nil {
		return []string{fmt.Sprintf("pi install failed: %v — run '%s' manually, then re-run 'pigo install'", err, cmd)}
	}

	// pi is now installed — chain straight into extension setup so the user
	// doesn't have to re-run just for that step.
	summary := []string{fmt.Sprintf("installed pi via npm (%s)", piNpmPkg)}
	return append(summary, installPiExtensions(piExtDir)...)
}
