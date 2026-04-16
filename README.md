# pigo

A resident AI knowledge server in Go. Ships as a single binary with automatic integration for [pi](https://github.com/nicholasgasior/pi) and [Claude Code](https://claude.ai/claude-code).

Pull the repo, build, run. Your AI gets persistent, semantic, versioned memory — not flat files.

## Quick Start

```bash
# Build
make build

# Set up vault + detect & install integrations
./pigo install

# Write a note
./pigo vault write --title "Architecture Decision" --tags "go,architecture" --body "We chose Go for single-binary deployment."

# Search
./pigo vault search "architecture"

# Start the server (for real-time extension support)
./pigo serve
```

Or install from source if you have Go:

```bash
go install github.com/scotmcc/pigo/cmd/pigo@latest
pigo install
```

## What It Does

Most AI memory is disposable. The session opens, the AI reads some context, the session closes. Next time, it starts from zero.

pigo is different. It runs a vault of markdown notes, indexed for semantic search, versioned with git, with facts extracted automatically over time. The AI doesn't just query the knowledge — it *builds* it.

**Three ways to use it:**

| Interface | Who | How |
|-----------|-----|-----|
| **CLI** | Humans, scripts | `pigo vault search "auth patterns"` |
| **Pi extension** | Pi harness users | Native AI tools via persistent TCP pipe |
| **Claude Code skill** | Claude Code users | AI learns to call `pigo` via bash |

`pigo install` detects what's on your machine and sets up the right integration automatically.

## The Stack

```
Markdown files        — truth layer. Human-readable, editable anywhere.
SQLite + sqlite-vec   — search index. Fuzzy on title/tags, semantic on chunks.
go-git                — history. Every write is a commit. Full versioning.
Ollama                — embeddings. Local, no external API dependency.
```

One binary. No Docker. No Postgres. No cloud dependency.

## Requirements

- **Go 1.21+** (to build from source)
- **Ollama** running locally (optional — fuzzy search works without it, semantic search needs embeddings)
- A C compiler (for SQLite CGO — comes with Xcode on macOS, `gcc` on Linux)

## The Vault

```
vault.read   — read a note by title or path
vault.write  — create a note; auto-chunks, embeds, commits
vault.edit   — update a note; re-indexes changed chunks, commits
vault.search — fuzzy on title/tags + semantic on content chunks
vault.list   — browse everything in the vault
```

On every write, the file is split on headings. Each chunk is vectorized and stored in SQLite. Search returns ranked results with jump targets — not just "this file" but "this section of this file."

Notes are plain markdown with YAML frontmatter. Edit them in any editor. The vault is a git repo — every change is committed automatically.

## Server

```bash
pigo serve
```

Starts an HTTP API (port 9876) and a TCP pipe server (port 9877). The HTTP API accepts JSON commands. The TCP pipe supports persistent connections with async streaming — this is what makes the pi extension work in real-time.

```bash
# Health check
curl http://localhost:9876/health

# Send a command
curl -X POST http://localhost:9876/command \
  -d '{"command": "vault.search", "args": {"q": "architecture"}}'

# Discover all commands
curl -X POST http://localhost:9876/command \
  -d '{"command": "system.methods"}'
```

## Configuration

`~/.pigo/config.toml` (created by `pigo install`):

```toml
[vault]
path = "~/.pigo/vault"

[db]
path = "~/.pigo/pigo.db"

[ollama]
endpoint = "http://localhost:11434"
model = "nomic-embed-text"

[server]
host = "127.0.0.1"
port = 9876
pipe_port = 9877
```

All fields have sensible defaults. Zero config gets you a working vault.

## Architecture

Three layers. Nothing skips a layer.

```
Layer 3 — Commands     Thin handlers. Register via init(). Call layer 2.
Layer 2 — Business     Compose layer-1 calls. No direct external I/O.
Layer 1 — Integrations One thing each. SQLite, git, Ollama, HTTP.
```

See [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) for the full blueprint.

## Diagnostics

```bash
pigo about
```

Shows build info, config paths, detected integrations, service status, and vault stats.

## Docs

- [ARCHITECTURE.md](docs/ARCHITECTURE.md) — layers, packages, data model, wire format
- [ROADMAP.md](docs/ROADMAP.md) — phased build plan with exit criteria
- [CHECKLIST.md](docs/CHECKLIST.md) — granular task list by phase
- [ideas/](docs/ideas/) — future feature concepts

## License

[MIT](LICENSE)
