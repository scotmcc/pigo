# Checklist

Working document. Check items as they're completed.

---

## Phase 0 — Project Scaffold [DONE]

- [x] Initialize `go.mod` (`github.com/scotmcc/pigo`)
- [x] Create directory structure: `cmd/pigo/`, `internal/{config,db,git,ollama,vault,facts,server,commands,keys}/`
- [x] Create `extensions/pi/` directory
- [x] Write `internal/keys/keys.go` — command names, config keys, injectable build metadata
- [x] Write `internal/config/config.go` — TOML loading, defaults
- [x] Write `cmd/pigo/main.go` — Cobra root command, version subcommand
- [x] Write `.gitignore`
- [x] Verify: `go build ./cmd/pigo` compiles clean
- [x] Verify: `./pigo version` prints version string

---

## Phase 1 — The Vault [DONE]

### Layer 1: SQLite (`internal/db/`)

- [x] SQLite connection wrapper (open, close, WAL pragma)
- [x] Schema migration — create `notes`, `chunks`, and `facts` tables
- [ ] Load sqlite-vec extension (embeddings stored but vec_distance_cosine not yet available)
- [x] Note CRUD — insert, update, get by id, get by path, list, delete
- [x] Chunk CRUD — insert batch, delete by note_id, get by note_id
- [x] Vector search query — cosine similarity via sqlite-vec (ready, needs extension loaded)
- [x] Fuzzy search query — LIKE on title and tags, return ranked notes
- [x] Verify: can create DB, run migrations, insert/query notes and chunks

### Layer 1: Git (`internal/git/`)

- [x] Init repo (or open existing) at vault path
- [x] Add file to index
- [x] Commit with message — `vault: <title> [write|edit]`
- [x] Log commits for a file path
- [x] Verify: vault directory becomes a git repo, commits appear in log

### Layer 1: Ollama (`internal/ollama/`)

- [x] HTTP client for `/api/embed` endpoint
- [x] Configurable endpoint URL and model name
- [x] Takes string, returns `[]byte` (float32 vector as raw bytes)
- [x] Verify: embedding a sentence returns a vector of expected dimension (768-dim, 3072 bytes)

### Layer 2: Vault Operations (`internal/vault/`)

- [x] Frontmatter parser — YAML frontmatter ↔ Go struct
- [x] Markdown chunker — split on headings, generate anchor slugs, handle intro chunk
- [x] Slug generator — title → filesystem-safe slug
- [x] **vault.write** — save file, parse, chunk, embed, index, commit
- [x] **vault.read** — resolve by id/slug/path, return frontmatter + body
- [x] **vault.edit** — load existing, update body/tags, re-chunk, re-embed, re-index, commit
- [x] **vault.search** — run fuzzy + semantic (graceful when Ollama down), merge and rank
- [x] **vault.list** — list all notes with title, tags, date
- [ ] Relationship generation — after write, search for similar notes, populate `relates_to`

### Layer 3: CLI Commands

- [x] `pigo vault write --title "..." --tags "a,b" --body "..."` (+ `--stdin`)
- [x] `pigo vault read <id>`
- [x] `pigo vault edit <id> --body "..." --tags "a,b"`
- [x] `pigo vault search <query>` — print ranked results with scores
- [x] `pigo vault list` — list all notes (title, tags, date)
- [x] Verify: full write → search → read → edit cycle works end-to-end

---

## Phase 2 — Server + Command Registry [DONE]

### Command Registry

- [x] Registry type — map of command name → handler
- [x] `init()` auto-registration pattern in each command file
- [x] Command interface with `SendFunc` callback for progress streaming
- [x] `system.methods` command — returns registered commands with descriptions
- [x] `system.ping` — async test command (3 updates with delays)
- [x] Sync vs Async flag (`Info{Async: bool}`) on registration

### Server (`internal/server/`)

- [x] HTTP listener — REST endpoints wrapping command dispatch
- [x] Request/response wire format with GUID
- [x] Graceful shutdown on SIGTERM/SIGINT
- [x] Health check endpoint

### Persistent Pipe (`internal/server/pipe.go`)

- [x] TCP listener on configurable port (default 9877)
- [x] Pipe registration — `{"type": "register_pipe"}`
- [x] GUID-based routing for async responses
- [x] Sync commands respond inline, async commands ack + stream updates
- [x] Newline-delimited JSON protocol

### CLI

- [x] `pigo serve` — starts HTTP + TCP pipe servers
- [ ] Client mode detection — auto-route through running server (not yet, direct mode only)

---

## Phase 3 — Pi Extension + Integrations [DONE]

### Pi Extension (`extensions/pi/pigo.ts`)

- [x] Persistent TCP pipe connection to daemon
- [x] GUID routing for async command responses
- [x] Steer messages for async progress updates
- [x] CLI fallback when daemon not running
- [x] System prompt injection from `~/.pigo/system.md`
- [x] Tools: `vault_search`, `vault_read`, `vault_write`, `vault_edit`, `vault_list`, `pigo_command`
- [x] Slash command: `/vault <query>`
- [x] Installation README

### Ollama Extension (`extensions/pi/ollama.js`)

- [x] Registers Ollama as a model provider in pi
- [x] Reads host from pigo config, falls back to `localhost:11434`
- [x] Filters to tool-capable models only
- [x] Slash commands: `/ollama-list`, `/ollama-ps`, `/ollama-show`, `/ollama-pull`, `/ollama-delete`, `/ollama-config`
- [x] AI tools: `ollama-list-models`, `ollama-running-models`, `ollama-pull-model`

### Claude Code Skill (`skills/claude/pigo.md`)

- [x] Skill file teaching Claude to use pigo CLI
- [x] Command reference with examples
- [x] Usage guidelines (search before writing, tag consistently, etc.)

### Install System

- [x] `pigo install` — creates `~/.pigo/`, detects pi + Claude, writes appropriate files
- [x] `pigo uninstall` — removes integrations, preserves vault data
- [x] `pigo upgrade` — alias for install (idempotent)
- [x] All assets embedded in binary via `//go:embed`

---

## Build + Release [DONE]

- [x] `Makefile` with ldflags for version/commit/date injection
- [x] `pigo version` — shows version, commit, build date
- [x] `pigo about` — full diagnostics (build, paths, integrations, services, vault stats)
- [x] `.goreleaser.yml` — cross-platform release config
- [x] `.github/workflows/ci.yml` — build + vet + test on push
- [x] `.github/workflows/release.yml` — GoReleaser on tag push
- [x] `LICENSE` (MIT)
- [x] README with quick start, architecture, config reference

---

## v0.1.0 Release Readiness [IN PROGRESS]

### Must Have

- [ ] **Tests** — vault core (write/read/edit/search), chunking, frontmatter, slug generation
- [ ] **Auto-init** — first vault command auto-creates `~/.pigo/`, runs migrations, just works
- [ ] **Graceful Ollama-down** — clear message ("semantic search disabled"), fuzzy still works
- [ ] **`--json` output** — structured JSON output on vault commands for scripting/extensions
- [ ] **Better errors** — "note not found", "Ollama not running", no stack traces

### Should Have

- [ ] sqlite-vec extension loading for real semantic search
- [ ] Relationship generation (`relates_to` in frontmatter)
- [ ] Example workflow in README

---

## Future Phases (not started)

### Phase 4 — Fact Extraction

- [ ] Extraction prompt + LLM inference via Ollama `/api/generate`
- [ ] Response parser (JSON → fact structs)
- [ ] Incremental extraction (track last extraction date)
- [ ] CLI: `pigo facts consolidate`, `pigo facts search`, `pigo facts topics`
- [ ] Extension tools: `fact_search`, `fact_topics`

### Phase 5 — Web Import

- [ ] `internal/fetch/` — HTTP client + HTML→markdown converter
- [ ] `vault.import` command — fetch URL, convert, save as vault note
- [ ] CLI: `pigo vault import <url>`

### Phase 6 — Job System

- [ ] Cron framework, job runner goroutine
- [ ] System jobs (nightly consolidation) + dynamic jobs (user-created)
- [ ] Job lifecycle: create, pause, resume, cancel
- [ ] CLI: `pigo jobs list/create/pause/resume/cancel/run`

---

## Decisions (Resolved)

- **Project name:** pigo (pi = the coding harness, go = the language)
- **Module path:** `github.com/scotmcc/pigo` (local git for now, push when ready)
- **Data directory:** `~/.pigo/` — vault/, pigo.db, config.toml
- **Go version:** 1.25.6 (latest stable)
- **Ollama model:** `nomic-embed-text` default, configurable
- **Config format:** TOML
- **CLI framework:** Cobra
- **License:** MIT
- **Distribution:** GitHub Releases via GoReleaser + `go install`
