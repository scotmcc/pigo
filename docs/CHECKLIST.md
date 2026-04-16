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
- [x] Relationship generation — after write, search for similar notes, populate `relates_to` (see Phase 3c)

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

- [x] TCP listener on configurable port (default 14160)
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

## Phase 3c — Knowledge Graph [DONE]

### Auto-relationships (`internal/vault/relations.go`)

- [x] After each write, search vault with note body as query, top N similar notes become `relates_to`
- [x] `maxRelations` cap (default 5) to keep frontmatter readable
- [x] Self-exclusion so a note doesn't relate to itself
- [x] Re-commits frontmatter with a `[relations]` tag in the commit message

### Wiki links (`internal/vault/wikilinks.go`)

- [x] Regex detection of `[[slug]]` and `[[slug|display]]` syntax
- [x] Slugification + deduplication of detected links
- [x] Stored as `links_to` in frontmatter
- [x] Unit tests in `wikilinks_test.go`

### Backlinks + combined Links API

- [x] `Service.Backlinks(noteID)` — reverse lookup across `relates_to` and `links_to`
- [x] `Service.Links(noteID)` — unified `LinksInfo` struct with all three link types
- [x] Registered command `vault.links`, CLI `pigo vault links <id>`

### Tags (`internal/vault/tags.go`)

- [x] `Service.Tags()` — aggregate tag counts across the vault, sorted by frequency
- [x] Registered command `vault.tags`, CLI `pigo vault tags`

---

## Phase 3d — Soul System [DONE]

- [x] `internal/assets/prompts/soul_preamble.md` — identity prompt for AI sessions
- [x] `internal/assets/prompts/welcome.md` — first-run greeting for new users
- [x] Assets embedded via `internal/assets/assets.go`
- [x] `cmd/pigo/soul.go` — CLI command to display soul content
- [x] `internal/commands/soul.go` — registered `soul` command for extensions
- [x] pi extension pulls soul preamble on `before_agent_start` (with `~/.pigo/system.md` override still honored)

---

## Phase 5 — Web Fetch + Import [DONE]

### Layer 1 (`internal/fetch/`)

- [x] HTTP client fetches a URL, returns HTML
- [x] HTML → markdown converter (regex-based, no DOM parser)
- [x] Unit tests in `markdown_test.go`

### Layer 2 (`internal/vault/import.go`)

- [x] `vault.import` business logic — fetch, convert, save as vault note
- [x] Source URL stored in frontmatter
- [x] Auto-tags with domain name and `imported` tag

### Layer 3

- [x] `pigo vault import <url>` (`cmd/pigo/vault_import.go`)
- [x] Registered command `vault.import` for extensions (`internal/commands/vault_import.go`)

---

## Phase 5b — Web Search [DONE]

- [x] `internal/search/search.go` — SearXNG HTTP client
- [x] Configurable SearXNG endpoint in `~/.pigo/config.toml`
- [x] Registered command `web.search` (`internal/commands/web_search.go`, `web_deps.go`)
- [x] `pigo web search <query>` CLI (`cmd/pigo/web.go`, `cmd/pigo/web_search.go`)
- [x] Exposed to pi extension as a native AI tool

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

### Must Have — Feature Polish

- [x] **Tests** — vault core (`vault_test.go`), chunking (`chunk_test.go`), frontmatter (`frontmatter_test.go`), slug generation (`slug_test.go`), wiki links (`wikilinks_test.go`), web fetch (`markdown_test.go`)
- [x] **Auto-init** — `cmd/pigo/setup.go` auto-creates `~/.pigo/` and vault dir on first use, runs migrations, prints first-run message
- [x] **Graceful Ollama-down** — `internal/vault/search.go` falls back to fuzzy with warning when Ollama is unreachable
- [x] **`--json` output** — `--json` persistent flag honored by all vault commands, `soul`, and `web search`
- [ ] **Better errors audit** — spot checks show errors are wrapped with `%w` and no stack traces leak, but do a full pass to confirm user-facing messages are clear ("note not found", "Ollama not running", etc.)

### Must Have — Install Automation

> **Principle:** the install should not require the user to "figure anything out." Detect what's present, ask permission before doing anything heavy or system-changing, and if the user declines — tell them the exact command and how to re-run `pigo install`.

- [x] **sqlite-vec bundled** — `asg017/sqlite-vec-go-bindings/cgo` statically linked; `init()` in `internal/db/db.go` calls `sqlite_vec.Auto()` so every connection has the extension. Two tests in `db_test.go` verify `vec_version()` and `vec_distance_cosine`.
- [x] **Ollama model pull (Ollama present)** — `cmd/pigo/install_ollama.go::ollamaStep`. Detects on-PATH + reachable, asks before pulling, prints exact command on decline, streams `ollama pull` output on accept.
- [x] **Ollama absent** — `offerOllamaInstall` in `install_ollama.go`. `brew install ollama` on macOS, `curl -fsSL https://ollama.com/install.sh | sh` on Linux, download-page fallback otherwise. Declines and failures both print the exact command and re-run hint.
- [x] **pi detection + install** — `cmd/pigo/install_pi.go::piStep`. On-PATH or extensions dir → install extensions. Absent + npm → offer `npm install -g @mariozechner/pi-coding-agent`, auto-chain to extension install on success. npm absent → point at nvm/brew with exact commands.
- [x] **Claude Code detection** — unchanged (file-drop, no external to install). `pigo doctor` surfaces status. No install-time prompt needed.
- [x] **`pigo doctor`** — `cmd/pigo/doctor.go`. Reports sqlite-vec / Ollama / embedding model / pi extension / Claude skill. Actionable fix per red row. Exit 1 on any failure. Supports `--json` for scripting.

### Should Have

- [x] Relationship generation (`relates_to` in frontmatter) — see Phase 3c
- [x] Example workflow in README — Quick Start covers write → search → serve
- [ ] sqlite-vec extension loading — folded into Install Automation above

### Should Have

- [ ] sqlite-vec extension loading for real semantic search
- [ ] Relationship generation (`relates_to` in frontmatter)
- [ ] Example workflow in README

---

## Future Phases

### Phase 4 — Fact Extraction [DEFERRED — under review]

> The knowledge graph (Phase 3c) already covers connection discovery and topic clustering at the note level. Do not start this work without a design review — confirm what facts would add on top of the graph first.

- [ ] Extraction prompt + LLM inference via Ollama `/api/generate`
- [ ] Response parser (JSON → fact structs)
- [ ] Incremental extraction (track last extraction date)
- [ ] CLI: `pigo facts consolidate`, `pigo facts search`, `pigo facts topics`
- [ ] Extension tools: `fact_search`, `fact_topics`

### Phase 6 — Jobs (Agent-Spawned Background Work) [NEXT]

**Design notes:** Unified model — one jobs table, two triggers (run-now and cron). pigo is agent-agnostic: a job is argv + a prompt. Orchestrator patterns (research/plan/implement/review) live in the prompt, not in code. Prompts-as-vault-notes is deferred pending real usage.

### Layer 1

- [ ] `jobs` table migration: id, prompt, command, args, status, created_at, started_at, finished_at, exit_code, output_path, schedule, next_run_at
- [ ] Cron expression parser integration (library choice TBD)

### Layer 2 (`internal/jobs/`)

- [ ] Launcher — fork/exec target, stream stdout+stderr to `~/.pigo/jobs/<id>.log`, return job id immediately
- [ ] Supervisor — watch running processes, detect exit, update status + exit code
- [ ] Orphan recovery — on `pigo serve` restart, reconcile previously-running jobs
- [ ] Scheduler loop — evaluate cron expressions, call launcher when a job is due
- [ ] Notifier — emit state-change events (`started`, `finished`, `failed`) over the pipe

### Layer 3

- [ ] `pigo jobs run --prompt "..." --target <claude|pi|bash>`
- [ ] `pigo jobs run ... --at <time>` (future-dated one-off)
- [ ] `pigo jobs run ... --cron <expr>` (recurring)
- [ ] `pigo jobs list`
- [ ] `pigo jobs status <id>`
- [ ] `pigo jobs output <id> [--tail N]`
- [ ] `pigo jobs stop <id>`
- [ ] `pigo jobs delete <id>`
- [ ] Registered commands for extensions: `jobs.run`, `jobs.list`, `jobs.status`, `jobs.output`, `jobs.stop`, `jobs.delete`

### Deferred (decide later)

- [ ] Prompts-as-vault-notes — revisit after real orchestrator prompts exist
- [ ] Built-in orchestrator types (research/plan/implement/review as first-class in pigo)

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
