package main

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/scotmcc/pigo/internal/config"
	"github.com/scotmcc/pigo/internal/detect"
)

// ollamaStep inspects the Ollama situation and returns a one-line summary
// describing what it did (or didn't do) for the final install report.
// The prompt and pull output, if any, stream directly to stdout.
func ollamaStep(cfg config.Config) string {
	if !detect.OllamaOnPath() {
		return offerOllamaInstall()
	}

	if !detect.OllamaReachable(cfg.Ollama.Endpoint) {
		return fmt.Sprintf("ollama is installed but not reachable at %s — start it with 'ollama serve' and re-run 'pigo install'", cfg.Ollama.Endpoint)
	}

	has, err := detect.OllamaHasModel(cfg.Ollama.Endpoint, cfg.Ollama.Model)
	if err != nil {
		return fmt.Sprintf("ollama model check failed: %v — continuing", err)
	}

	if has {
		return fmt.Sprintf("ollama ready — model %q already present", cfg.Ollama.Model)
	}

	fmt.Println()
	q := fmt.Sprintf("pigo needs to download the embedding model %q for semantic search. Proceed?", cfg.Ollama.Model)
	if !confirm(q, false) {
		return fmt.Sprintf("skipped model pull — run 'ollama pull %s' then re-run 'pigo install'", cfg.Ollama.Model)
	}

	fmt.Printf("\npulling %s...\n", cfg.Ollama.Model)
	if err := ollamaPull(cfg.Ollama.Model); err != nil {
		return fmt.Sprintf("model pull failed: %v — run 'ollama pull %s' manually", err, cfg.Ollama.Model)
	}
	return fmt.Sprintf("pulled ollama model %q", cfg.Ollama.Model)
}

// ollamaPull runs `ollama pull <model>` with its output streamed to the
// terminal so the user sees real-time progress.
func ollamaPull(model string) error {
	c := exec.Command("ollama", "pull", model)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	return c.Run()
}

// offerOllamaInstall is the "Ollama absent" branch of the ask-then-offer
// pattern. Picks a platform-appropriate install command if one is available,
// asks the user, and either runs it or tells them the exact command to run
// themselves — never leaving them to figure it out.
func offerOllamaInstall() string {
	cmd, label := ollamaInstaller()
	if cmd == "" {
		return "ollama not installed — download from https://ollama.com/download, then re-run 'pigo install'"
	}

	fmt.Println()
	fmt.Printf("ollama is not installed. pigo can install it via %s:\n  %s\n", label, cmd)
	if !confirm("Run this now?", false) {
		return fmt.Sprintf("skipped ollama install — run '%s' yourself, then re-run 'pigo install'", cmd)
	}

	fmt.Printf("\nrunning: %s\n\n", cmd)
	if err := runShell(cmd); err != nil {
		return fmt.Sprintf("ollama install failed: %v — run '%s' manually, then re-run 'pigo install'", err, cmd)
	}
	return "installed ollama — if it's not already running, start it ('ollama serve' or via your service manager) and re-run 'pigo install' to pull the embedding model"
}

// ollamaInstaller returns the platform-appropriate install command and a
// short label identifying the installer. Empty strings mean no auto-install
// option is available on this platform — we'll just point at the download page.
func ollamaInstaller() (cmd, label string) {
	switch {
	case detect.IsMac() && detect.OnPath("brew"):
		return "brew install ollama", "Homebrew"
	case detect.IsLinux() && detect.OnPath("curl"):
		return "curl -fsSL https://ollama.com/install.sh | sh", "the official install script"
	}
	return "", ""
}

// runShell runs a shell command with stdio passed through so the user can
// interact with any prompts (sudo, homebrew taps, etc.).
func runShell(cmd string) error {
	c := exec.Command("sh", "-c", cmd)
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	return c.Run()
}
