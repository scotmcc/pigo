# Roadmap

Milestone-driven, not date-driven. Each phase delivers something usable. Later phases depend on earlier ones but each phase stands on its own.

**Status:** **v0.1.0 shipped** (tag pushed 2026-04-16). Linux release binaries via GoReleaser. macOS users install via `go install`. Phases 0–3 complete plus 3c (knowledge graph), 3d (soul system), 5 (web import), 5b (web search). Phase 4 (facts) deferred pending design review. Next focus: **post-v0.1.0 hardening** (below) based on external review — provenance, prompt-injection defense, embedding model pinning, macOS binaries, pitch/positioning fixes. Jobs (Phase 6) follows.

---

## Post-v0.1.0 — Honesty, Security, Provenance **[NEXT]**

**Delivers:** A tighter, more honest, more auditable v0.2. Everything in this bucket came out of external review of the v0.1.0 repo and the original positioning doc. None of it is a research problem; all of it is shippable in incremental commits.

**What gets built:**

### Provenance and vault quality
- **Source frontmatter field** — `source: human | ai | imported | external` on every note. Default stamped by whichever path creates the note (`vault.write` called by AI → `ai`; `vault.import` → `imported`; manual file creation → `human`).
- **`--source` filter on search** — `pigo vault search q --source=human` lets a user pull only curated content when context is tight. Cheapest hedge against the "vault rot" failure mode.
- **`pigo vault gc`** — optional garbage-collect pass: detect near-duplicate notes for review, optionally squash trivial AI-only commit churn. Doesn't solve AI-write-quality; gives the user a periodic cleanup lever.

### Security
- **Import safety** — `vault.import` stamps `source: external` and adds a prompt guideline treating imported content as data not instructions. Closes the prompt-injection path where a malicious URL becomes trusted context on later searches.
- **Threat model in README** — shipped 2026-04-16 as part of this push: localhost-only, no auth, machine is the trust boundary, binding beyond localhost means you own the auth problem.

### Correctness
- **Embedding model pinning** — store the model name (and ideally a hash) in the chunks table. On `pigo serve` / `pigo install`, detect mismatch with the configured model and offer to reindex. Prevents silent search degradation on model upgrade.

### Distribution
- **macOS release binaries** — matrix runner on `macos-latest` in `.github/workflows/release.yml` so Mac users don't need Go + CGO just to try pigo. Biggest remaining install-friction win.

### Discoverability and voice
- **`pigo examples` command** — printable cheat sheet of common workflows. Low-effort win for newcomers.
- **Pitch/README voice alignment** — README's "here's what it is, here are the commands, stop" voice carried through any external pitch text. README updated in this push; positioning doc/external pitch is a follow-up.
- **Positioning for compliance-constrained teams** — make explicit (in README or sibling doc) that pigo is the memory layer for shops that can't use cloud AI memory. This is the real wedge and the current doc doesn't say so.

### Larger origin story
- **Name the Gaia relationship** somewhere public-facing. "pigo is the memory layer extracted from a larger multi-agent system, packaged standalone for teams that need local AI memory" answers the sustainability question a reader will ask. Scot decides what to expose.

**Explicitly deferred to v0.3:**
- **Curation mode** — AI writes land in `pending/`, a human commits them before they enter the searchable vault. Addresses vault rot more fully but changes daily-use workflow and needs design work. Revisit after provenance + gc land and we see how the vault ages under real use.

**Dependencies:** Phase 1 (vault), Phase 5 (web import for source-marking), release workflow (for macOS binaries).

---

## Phase 0 — Project Scaffold **[DONE]**

**Delivers:** A buildable Go module with the right structure.

**What was built:**
- `go.mod`, directory structure, `internal/keys/`, `internal/config/`, `cmd/pigo/main.go`, `.gitignore`
- Cobra CLI with `--config` flag and `version` subcommand

---

## Phase 1 — The Vault **[DONE]**

**Delivers:** A working knowledge system. Create, read, edit, and search notes from the CLI.

**What was built:**
- Layer 1: `internal/db/` (SQLite), `internal/git/` (go-git), `internal/ollama/` (embeddings)
- Layer 2: `internal/vault/` — read, write, edit, search, list with chunking + embedding + git
- Layer 3: CLI subcommands for all vault operations
- Verified: full write → search → read → edit cycle works, embeddings stored, git history tracks changes

---

## Phase 2 — Server + Command Registry **[DONE]**

**Delivers:** A running daemon with HTTP API and a self-registering command registry.

**What was built:**
- `internal/commands/` — registry with init() auto-registration, system.methods
- `internal/server/` — HTTP server, JSON dispatch, health check, client
- `pigo serve` — starts daemon with graceful shutdown
- Verified: all vault commands work through HTTP API

---

## Phase 2b — Async Commands + Persistent Pipe **[DONE]**

**Delivers:** The infrastructure that makes pigo a real-time assistant backend, not just a request/response API. Commands can run in the background and stream progress updates.

This is the pattern proven in the original pigo — it's what makes the pi extension work well.

**What gets built:**

### Async Command Model
- `SendFunc` callback pattern — commands receive a `Send(status, message, data)` function they can call to emit progress updates without knowing about transport
- Sync vs Async flag on command registration — async commands return an immediate `accepted` ack, then stream updates through the pipe
- GUID-based routing — every command gets a unique ID, responses route back by GUID

### Persistent Pipe (TCP)
- TCP listener alongside the HTTP server (separate port or same port, protocol detection)
- Pipe registration — clients send `{"type":"register_pipe"}` to establish a persistent connection
- Newline-delimited JSON streaming — simple, no framing complexity
- Multiple commands in flight on one connection, routed by GUID

### Wire Format Upgrade
```json
// Request (gains GUID)
{ "guid": "abc-123", "command": "vault.search", "args": { "q": "..." } }

// Sync response (same as before)
{ "guid": "abc-123", "success": true, "data": [...] }

// Async ack (immediate)
{ "guid": "abc-123", "status": "accepted" }

// Async update (streamed)
{ "guid": "abc-123", "status": "update", "message": "Processing chunk 3/7..." }

// Async done
{ "guid": "abc-123", "status": "done", "data": {...} }

// Async error
{ "guid": "abc-123", "status": "error", "error": "..." }
```

**Exit criteria:**
- Async command streams progress updates through the pipe
- Multiple concurrent async commands route correctly by GUID
- Sync commands still work over HTTP unchanged
- Pipe reconnects cleanly if dropped

**Dependencies:** Phase 2.

---

## Phase 3 — Pi Extension **[DONE]**

**Delivers:** A TypeScript extension that makes pigo a native part of any pi coding session. The AI gets real tools, not bash wrappers.

This is the user-facing integration that makes pigo worth installing.

**What gets built:**

### Core Extension (`extensions/pi/pigo.ts`)
- Persistent TCP pipe to pigo daemon on startup
- GUID routing for async command responses
- Steer messages — async updates delivered to the AI mid-stream

### Tools Registered with the AI
- `vault_search` — semantic + fuzzy search the knowledge vault
- `vault_read` — read a specific note
- `vault_write` — create a new note (AI builds the knowledge graph)
- `vault_edit` — update an existing note
- `vault_list` — browse what's in the vault
- `pigo_command` — generic command executor for any registered command (extensible)

### System Prompt Injection
- On `before_agent_start`, inject a private system prompt from `~/.pigo/system.md`
- This teaches the AI about the user's vault, preferences, and pigo capabilities
- Users can customize what the AI knows about their setup without editing extension code

### Slash Commands
- `/vault` — quick vault search from the command line
- `/remember` — shortcut to write a note from the current conversation
- `/recall` — search vault and inject results into context

### Fallback Mode
- If pigo daemon isn't running, fall back to CLI execution via `pi.exec()`
- Graceful degradation — tools still work, just without streaming

**Exit criteria:**
- Extension loads in a pi session, connects to daemon
- AI can search vault and get results as native tool responses
- AI can write notes that persist across sessions
- Async commands stream progress to the AI as steer messages
- Extension works (degraded) even without a running daemon

**Dependencies:** Phase 2b (persistent pipe for full experience), Phase 1 (CLI for fallback).

---

## Phase 3b — Claude Code Skill **[DONE]**

**Delivers:** A skill file that teaches Claude Code to use pigo via CLI. No extension needed — just a markdown file that Claude reads.

**What gets built:**
- `skills/pigo.md` — skill definition that describes pigo CLI commands, when to use them, and output formats
- Installed by copying to `~/.claude/commands/` or project `.claude/commands/`
- Teaches Claude: "when the user mentions memory, vault, knowledge — use `pigo vault search/write/read`"

**Exit criteria:**
- Claude Code user can `/pigo search "auth patterns"` and get vault results
- Claude can autonomously decide to write a note when it learns something worth remembering
- Works without any daemon — pure CLI

**Dependencies:** Phase 1 (CLI must work).

---

## Phase 3c — Knowledge Graph **[DONE]**

**Delivers:** The vault becomes a real knowledge graph. Notes connect to each other automatically, explicitly, and in both directions.

**What was built:**

### Auto-discovered relationships (`internal/vault/relations.go`)
- After each write, the new note's body is used as a search query against the vault
- Top N semantically similar notes (default 5) become `relates_to` entries in the frontmatter
- The AI doesn't have to remember to link — similarity does it automatically

### Explicit wiki links (`internal/vault/wikilinks.go`)
- Obsidian-style `[[slug]]` and `[[slug|display]]` syntax detected in note bodies
- Referenced slugs stored as `links_to` in frontmatter
- Lets the AI (or a human editor) declare connections the graph should know about

### Backlinks + combined Links API (`internal/vault/relations.go`)
- `Backlinks(noteID)` — reverse lookup across `relates_to` and `links_to`
- `Links(noteID)` — returns `relates_to`, `links_to`, and `backlinks` together

### Tag aggregation (`internal/vault/tags.go`)
- `Tags()` — all tags across the vault with note counts, sorted by frequency

### CLI + command surface
- `pigo vault links <id>` — show all connections for a note
- `pigo vault tags` — browse tag cloud
- Registered commands: `vault.links`, `vault.tags`

**Dependencies:** Phase 1 (vault), Phase 2 (command registry).

---

## Phase 3d — Soul System **[DONE]**

**Delivers:** pigo is not an anonymous tool — it has an identity, a welcome flow, and a way to teach the AI about itself and the user. Replaces and expands on the narrow "System Prompt Injection" bullet from Phase 3.

**What was built:**

### Identity assets (`internal/assets/prompts/`)
- `soul_preamble.md` — the core identity/persona prompt injected into AI sessions
- `welcome.md` — the first-run greeting shown to new users

### Soul command (`cmd/pigo/soul.go`, `internal/commands/soul.go`)
- `pigo soul` — displays identity + welcome content
- Registered `soul` command available through the server/pipe for extensions to fetch

### System prompt injection
- pi extension pulls the soul preamble on `before_agent_start`
- Users can still override via `~/.pigo/system.md`; the soul preamble is the default

**Dependencies:** Phase 2 (command registry), Phase 3 (extension injection point).

---

## Phase 4 — Fact Extraction + Consolidation **[DEFERRED — under review]**

> **Status note (2026-04-16):** With the knowledge graph (Phase 3c) in place — auto-`relates_to`, wiki links, backlinks, tag aggregation — a large part of what "facts" was intended to deliver (topic clustering, cross-note connections) is already covered at the note level. Facts would still add structured extraction (importance scores, entities, queryable rows with source linkage), but whether that's worth the complexity depends on real usage patterns. This phase is paused pending a design review after the jobs system lands. Do not build this without revisiting scope first.

**Delivers:** The second layer of intelligence. Raw notes become structured facts. The knowledge compounds over time.

**What gets built:**

### Layer 1
- Fact storage in SQLite (facts table — already migrated)
- LLM inference client for extraction prompts (Ollama `/api/generate`)

### Layer 2
- `internal/facts/` — extraction service:
  - Takes a note, sends to LLM with extraction prompt
  - Parses structured JSON response (content, topic, importance, entities)
  - Stores facts with source linkage
- Incremental extraction — tracks last extraction date per note, only processes new/changed notes
- Consolidation pass — runnable as CLI command or triggered by job system

### Layer 3
- `pigo facts consolidate` — run extraction on new/changed notes
- `pigo facts search <query>` — search extracted facts
- `pigo facts topics` — list all fact topics
- Extension tools: `fact_search`, `fact_topics`

**Exit criteria:**
- `pigo facts consolidate` processes vault notes and extracts structured facts
- `pigo facts search "architecture decisions"` returns relevant facts with importance scores
- Running consolidate twice doesn't duplicate facts (incremental)
- Facts link back to source notes

**Dependencies:** Phase 1 (vault must exist to extract from).

---

## Phase 5 — Web Fetch + Import **[DONE]**

**Delivers:** Save web content directly into the vault. The AI can capture reference material from URLs without leaving the conversation.

**What was built:**

### Layer 1 (`internal/fetch/`)
- HTTP client that fetches a URL and returns HTML
- HTML → markdown converter (regex-based, no DOM parser) with tests

### Layer 2 (`internal/vault/import.go`)
- `vault.import` command — fetch URL, convert to markdown, save as vault note with source URL in frontmatter
- Auto-tag with domain name and "imported" tag

### Layer 3
- `pigo vault import <url>` — CLI command (`cmd/pigo/vault_import.go`)
- Registered command `vault.import` available over HTTP/pipe for extensions

**Dependencies:** Phase 1 (vault).

---

## Phase 5b — Web Search **[DONE]**

**Delivers:** The AI can search the open web from within a session, not just the vault. Combined with Phase 5, it can search → import → remember in one flow.

**What was built:**

### Layer 1 (`internal/search/`)
- HTTP client targeting a configurable SearXNG instance
- Endpoint URL configurable via `~/.pigo/config.toml`
- Self-hostable, no proprietary search API dependency

### Layer 2 / Layer 3
- Registered command `web.search` (`internal/commands/web_search.go`, `internal/commands/web_deps.go`)
- `pigo web search <query>` CLI (`cmd/pigo/web.go`, `cmd/pigo/web_search.go`)
- Available through the pi extension as a native tool

**Dependencies:** Phase 2 (command registry). Pairs with Phase 5 for search → import workflows.

---

## Phase 6 — Jobs (Agent-Spawned Background Work) **[NEXT]**

**Delivers:** A unified system for running long agent sessions outside the current conversation. One model, two triggers — "run now" (kicked off mid-session so the chat can continue) and "run on schedule" (cron recurring or future-dated). The user and the front-line AI stay lightweight and chatty while the real work happens in a background agent session.

**Design principle — pigo is agent-agnostic.** A job is *argv + a prompt*. pigo does not care whether the launched binary is `claude`, `pi`, `bash`, or anything else. The orchestrator pattern (research → plan → implement → review with sub-agents) lives entirely in the prompt, not in pigo. This keeps the infrastructure small and lets orchestration patterns evolve without schema or code changes.

**Typical flow:**
1. User and front-line AI chat through what should happen
2. AI crafts an orchestrator prompt
3. `pigo jobs run --prompt <...> --target claude` — returns a job id, keeps going
4. The orchestrator agent (in its own session) spawns research / planning / implementation / review sub-agents as it sees fit
5. pigo emits a "finished" event over the pipe
6. Front-line AI spot-checks the result and reports back

**What gets built:**

### Layer 1
- `jobs` table in SQLite: `id, prompt, command, args, status, created_at, started_at, finished_at, exit_code, output_path, schedule, next_run_at`
- Cron expression parser (library TBD — likely `robfig/cron`)

### Layer 2 (`internal/jobs/`)
- **Launcher** — fork/exec the target binary, stream stdout+stderr to `~/.pigo/jobs/<id>.log`, return the job id immediately
- **Supervisor** — watch running processes, detect exit, update status and exit code
- **Orphan recovery** — on `pigo serve` restart, reconcile processes that were running when the daemon stopped
- **Scheduler loop** — evaluate cron expressions, call the launcher when a job is due
- **Notifier** — emit async events over the persistent pipe (`started`, `finished`, `failed`) so extensions can wake up the AI when a job completes

### Layer 3
- `pigo jobs run --prompt "..." --target <claude|pi|bash> [--at <time>|--cron <expr>]`
- `pigo jobs list` — everything with status, schedule, next run
- `pigo jobs status <id>`
- `pigo jobs output <id> [--tail N]`
- `pigo jobs stop <id>`
- `pigo jobs delete <id>`
- Registered commands (`jobs.run`, `jobs.list`, `jobs.status`, `jobs.output`, `jobs.stop`, `jobs.delete`) so extensions can expose job management as native AI tools

**Exit criteria:**
- An agent session can launch a job, keep chatting, and get a pipe notification on completion
- Cron expression schedules a job that survives `pigo serve` restarts
- Launched process is reparented to the daemon — terminal crash / network drop does not kill it
- `pigo jobs list` shows live state, schedule, and next-run time
- Job failures are logged and surfaced, never crash the server

**Explicitly deferred (decide after we have real usage):**
- **Prompts-as-vault-notes.** Storing orchestrator prompts in the vault would give us versioning + knowledge-graph connections, but it raises an identity question — is a prompt a user-facing note or a system-level skill? And it would leak into vault search. Revisit once we have a few real orchestrator prompts written.
- **Built-in orchestrator types** (research / plan / implement / review as first-class concepts in pigo). Keep as prompt patterns for now.

**Dependencies:** Phase 2 (server + commands), Phase 2b (pipe for streaming notifications).

---

## Phase 7 — Context Awareness

**Delivers:** The AI knows more than just the vault — it knows about the current project, recent activity, and user context. This is what makes pigo feel like a real assistant, not a search box.

**What gets built:**

### Project Context
- `pigo context project` — scan current directory for project signals (package.json, go.mod, Cargo.toml, README, CLAUDE.md) and return a structured summary
- Extension injects project context into system prompt automatically

### Session Awareness
- Track what the AI has read/written/searched in the current session
- `pigo context session` — return session activity summary
- Useful for: "what have we been working on?" without re-reading everything

### User Preferences
- `~/.pigo/preferences.toml` — user-defined context the AI should always know
- Name, role, common projects, coding style preferences, timezone
- Injected into system prompt alongside vault context

### Conversation Memory
- On session end (or `/remember`), offer to save conversation highlights as vault notes
- Extension hook on `session_shutdown` to prompt for note creation

**Exit criteria:**
- AI knows the current project type and structure without being told
- AI can answer "what have we been working on?" from session context
- User preferences are reflected in AI behavior across sessions

**Dependencies:** Phase 3 (extension), Phase 1 (vault for storage).

---

## Phase 8 — Background Intelligence

**Delivers:** pigo thinks between sessions. It consolidates knowledge, discovers patterns, generates summaries — all without being asked.

**What gets built:**

### Nightly Consolidation (via job system)
- Extract facts from new/changed notes
- Re-evaluate `relates_to` across the vault based on new content
- Generate a "daily digest" note: what was added, what facts were extracted, what patterns emerged

### Relationship Discovery
- After consolidation, run a pass that finds notes with overlapping entities/topics
- Update `relates_to` in frontmatter and recommit
- The knowledge graph grows automatically

### Weekly Summary
- Weekly job that reviews all notes and facts from the past week
- Produces a summary note with themes, connections, and observations
- Stored in the vault — the system documents its own learning

**Exit criteria:**
- Nightly job runs, extracts facts, updates relationships
- Weekly summary captures themes and connections
- Vault grows richer over time without manual effort

**Dependencies:** Phase 4 (facts), Phase 6 (jobs).

---

## Phase 9+ — Future Ideas

These are ideas, not commitments. Each becomes its own phase when the time comes.

- **TTS / Speech** — Kokoro TTS integration. AI can speak responses. Voice selection by context. Reuse old pigo's Kokoro client.
- **Agents** — autonomous LLM entities for complex multi-step tasks (from Gaia.6)
- **system.exec** — run shell commands through the gateway with safety controls
- **Docker integration** — container management. Reuse old pigo's Docker-over-SSH pattern, upgrade to direct API.
- **Notifications** — push events to external systems (Slack, webhooks)
- **Web UI** — browser-based vault explorer served by pigo HTTP server
- **Remote sync** — scheduled git push to Gitea/GitHub
- **Multi-vault** — support multiple vaults (personal, work, project-specific)
- **Vault templates** — note templates for common types (decision, meeting, idea, person)
- **Semantic deduplication** — detect when new notes overlap significantly with existing ones

---

## Phase Dependency Map

```
Phase 0 (scaffold) ✓
  └── Phase 1 (vault) ✓
        ├── Phase 2 (server) ✓
        │     └── Phase 2b (async + pipe) ✓
        │           └── Phase 3 (pi extension) ✓
        │                 ├── Phase 3d (soul system) ✓
        │                 └── Phase 7 (context awareness)
        ├── Phase 3b (claude code skill) ✓
        ├── Phase 3c (knowledge graph) ✓
        ├── Phase 4 (facts) [DEFERRED]
        │     └── Phase 8 (background intelligence)
        ├── Phase 5 (web fetch + import) ✓
        │     └── Phase 5b (web search) ✓
        └── Phase 6 (jobs)  ← next
              └── Phase 8 (background intelligence)
```

Phases 3b, 3c, 3d, 5, and 5b have all landed in parallel with the core server work. Phase 6 (jobs) is the next intended focus. Phase 4 (facts) is on hold — see its section for the design-review note.

---

## What Makes This a Great AI Extension

The difference between a good tool and a great assistant:

1. **The AI builds knowledge, not just queries it.** Every conversation can produce notes. The vault grows from use.
2. **Knowledge compounds.** Facts are extracted, relationships form, summaries are generated. Yesterday's notes make today's answers better.
3. **It works between sessions.** Background jobs consolidate, discover, summarize. The AI picks up where it left off, but smarter.
4. **Context is automatic.** The AI knows the project, the user, the session history. No "let me re-explain everything."
5. **It degrades gracefully.** No daemon? CLI works. No Ollama? Fuzzy search works. No extension? Skill works. Every layer adds value but none is required.
6. **It's inspectable.** The vault is markdown files. The index is SQLite. The history is git. No black boxes.
