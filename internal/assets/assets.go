// Package assets embeds the pi extension and Claude Code skill into the binary.
// These files are written to disk by "pigo install".
//
// Go's embed directive bakes file contents into the compiled binary at build time.
// No external files needed at runtime — the binary is fully self-contained.
package assets

import _ "embed" // required for //go:embed directives

// PiExtension is the TypeScript extension for the pi coding harness.
//
//go:embed pi/pigo.ts
var PiExtension string

// OllamaExtension is the Ollama provider/management extension for pi.
//
//go:embed pi/ollama.js
var OllamaExtension string

// ClaudeSkill is the markdown skill file for Claude Code.
//
//go:embed claude/pigo.md
var ClaudeSkill string

// WelcomePrompt is shown to the AI when no soul file exists.
// It guides the AI through getting to know the user.
//
//go:embed prompts/welcome.md
var WelcomePrompt string

// SoulPreamble is prepended to the soul file content when injecting
// into the system prompt. It explains pigo's role and tools.
//
//go:embed prompts/soul_preamble.md
var SoulPreamble string
