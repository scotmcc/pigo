# pigo

A resident AI knowledge server in Go. Ships as a single binary with automatic integration for [pi](https://pi.dev) and [Claude Code](https://claude.ai/claude-code).

Pull the repo, build, run. Your AI gets persistent, semantic, versioned memory — not flat files.

## Quick Start

```bash
# Build
make build

# Set up vault + detect & install integrations
./pigo install

# Verify everything is wired up (green/red per dependency)
./pigo doctor

# Write a note
./pigo vault write --title "Architecture Decision" --tags "go,architecture" --body "We chose Go for single-binary deployment."

# Search
./pigo vault search "architecture"

# Start the server (for real-time extension support)
./pigo serve

# Stop a running server
./pigo stop

# Check daemon + dependency status later
./pigo status
./pigo doctor

# Re-check integrations, prompt to upgrade pi if a newer one is out
./pigo upgrade
```

Or install from source if you have Go:

```bash
go install github.com/scotmcc/pigo/cmd/pigo@latest
pigo install
```

## What It Does

Most AI chat sessions start from zero. You explain your project, your preferences, your history; next session you explain them again. Cloud AI memory features help, but they're proprietary, vendor-locked, and off-limits for any code under real compliance constraints.

pigo runs a local vault of markdown notes, indexed for semantic search and versioned with git. Any AI agent with pigo tools can read from and write to it. The knowledge stays on your machine as plain files — inspectable, portable, and unaffected by a vendor's product decisions.

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

- **Ollama** running locally with an embedding model pulled. This is effectively required — fuzzy search works without it, but the "AI builds the knowledge" story assumes semantic retrieval. `pigo install` detects Ollama and offers to pull `nomic-embed-text` automatically.
- To build from source: **Go 1.21+** and a C compiler (sqlite-vec needs CGO — Xcode on macOS, `gcc` on Linux). Prebuilt Linux binaries are shipped via GitHub Releases; macOS prebuilt binaries are coming (see ROADMAP).

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

Starts an HTTP API (port 14159) and a TCP pipe server (port 14160). The HTTP API accepts JSON commands. The TCP pipe supports persistent connections with async streaming — this is what makes the pi extension work in real-time.

```bash
# Health check
curl http://localhost:14159/health

# Send a command
curl -X POST http://localhost:14159/command \
  -d '{"command": "vault.search", "args": {"q": "architecture"}}'

# Discover all commands
curl -X POST http://localhost:14159/command \
  -d '{"command": "system.methods"}'
```

### Trust model

pigo binds to `127.0.0.1` and has no authentication. The trust boundary is the local machine — anything with shell access to your user can already read the vault directly, so HTTP/pipe add nothing for an attacker already inside. If you change `host` in `config.toml` to bind beyond localhost, you've taken on the auth problem yourself; pigo doesn't try to solve it.

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
port = 14159
pipe_port = 14160
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
