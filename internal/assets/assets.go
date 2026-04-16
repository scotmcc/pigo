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
