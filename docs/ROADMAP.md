# Roadmap

Milestone-driven, not date-driven. Each phase delivers something usable. Later phases depend on earlier ones but each phase stands on its own.

**Status:** Phases 0–3 complete (vault, server, async pipe, pi extension, ollama extension, claude skill, install system, build/release pipeline). Working on v0.1.0 release readiness (tests, auto-init, polish).

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

## Phase 4 — Fact Extraction + Consolidation

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

## Phase 5 — Web Fetch + Import

**Delivers:** Save web content directly into the vault. The AI can capture reference material from URLs without leaving the conversation.

**What gets built:**

### Layer 1
- `internal/fetch/` — HTTP client that fetches a URL and returns HTML
- HTML → markdown converter (regex-based, like old pigo — no DOM parser needed)

### Layer 2
- `vault.import` command — fetch URL, convert to markdown, save as vault note with source URL in frontmatter
- Configurable: keep raw HTML, convert to markdown, or both
- Auto-tag with domain name and "imported" tag

### Layer 3
- `pigo vault import <url>` — CLI command
- Extension tool: `vault_import` — AI can save web pages as notes

**Exit criteria:**
- `pigo vault import https://example.com/article` creates a well-formatted vault note
- Imported note has proper frontmatter with source URL
- Extension can import URLs during a conversation

**Dependencies:** Phase 1 (vault).

---

## Phase 6 — Job System

**Delivers:** Background task scheduling. The server can run recurring work automatically.

**What gets built:**

### Layer 1
- Cron expression parser (library or built-in)
- Job storage in SQLite

### Layer 2
- `internal/jobs/` — job runner:
  - System jobs (always running): nightly consolidation, relationship refresh
  - Dynamic jobs (user-created): custom schedules wrapping any registered command
  - Job lifecycle: create, pause, resume, cancel
  - Execution tracking: last run, next run, last error

### Layer 3
- `pigo jobs list` — show all jobs with status
- `pigo jobs create --name "..." --cron "..." --command "..."`
- `pigo jobs pause/resume/cancel <id>`
- `pigo jobs run <name>` — execute immediately
- Extension tools for job management

**Exit criteria:**
- Server runs nightly consolidation automatically
- `pigo jobs list` shows scheduled jobs with next execution time
- Jobs survive server restart
- Job failures are logged, don't crash the server

**Dependencies:** Phase 2 (server), Phase 4 (consolidation is the first real job).

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
        │     └── Phase 2b (async + pipe)
        │           └── Phase 3 (pi extension)
        │                 └── Phase 7 (context awareness)
        ├── Phase 3b (claude code skill) — can start anytime after Phase 1
        ├── Phase 4 (facts)
        │     └── Phase 8 (background intelligence)
        ├── Phase 5 (web fetch + import)
        └── Phase 6 (jobs)
              └── Phase 8 (background intelligence)
```

Phases 3b, 4, and 5 can run in parallel — they're independent of each other.

---

## What Makes This a Great AI Extension

The difference between a good tool and a great assistant:

1. **The AI builds knowledge, not just queries it.** Every conversation can produce notes. The vault grows from use.
2. **Knowledge compounds.** Facts are extracted, relationships form, summaries are generated. Yesterday's notes make today's answers better.
3. **It works between sessions.** Background jobs consolidate, discover, summarize. The AI picks up where it left off, but smarter.
4. **Context is automatic.** The AI knows the project, the user, the session history. No "let me re-explain everything."
5. **It degrades gracefully.** No daemon? CLI works. No Ollama? Fuzzy search works. No extension? Skill works. Every layer adds value but none is required.
6. **It's inspectable.** The vault is markdown files. The index is SQLite. The history is git. No black boxes.
