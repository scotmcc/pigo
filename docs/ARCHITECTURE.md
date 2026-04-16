# Architecture

pigo is a resident knowledge server in Go. It ships as a single binary with an optional pi harness extension. Users pull the repo, build, and get a working vault + CLI + server. The vault location is configurable — the repo is source code, not data.

---

## What pigo Is

An always-on personal knowledge system that supplements AI coding harnesses. Instead of flat markdown in `~/.claude/` or `.claude/`, the AI can query a real knowledge system — chunked, embedded, versioned, with relationships and facts extracted over time.

**Three ways to use it:**

| Interface | Who | How |
|-----------|-----|-----|
| CLI | Humans, scripts, skills | `pigo vault search "auth patterns"` |
| Server | Extensions, integrations | TCP/HTTP — always running, holds state |
| Pi extension | Pi harness users | JS extension registers native tools with the AI |

Claude Code users don't need the extension — a skill that teaches the AI to call `pigo` via bash is enough. Pi users get native tool registration via `pi.registerTool()`.

---

## Layer Architecture

Three layers. Nothing skips a layer.

```
Layer 3 — Commands / CLI
  Thin handlers. Register via init(). Call layer 2. Never call other commands.
  Examples: vault.read, vault.write, vault.search, system.methods

Layer 2 — Business Logic
  Compose layer-1 calls. No direct external I/O.
  Examples: write note + chunk + embed + index + commit

Layer 1 — External Integrations
  One thing each. No chaining, no business logic.
  Examples: SQLite, go-git, Ollama, HTTP client, file I/O
```

**The rule:** Layer 3 calls layer 2. Layer 2 calls layer 1. Layer 1 talks to the outside world. No skipping.

---

## Package Layout

```
pigo/
  cmd/
    pigo/
      main.go              — entry point, wires everything
  internal/
    config/                — configuration loading (TOML/YAML, env, defaults)
    db/                    — SQLite wrapper, schema migration, sqlite-vec setup
    git/                   — go-git wrapper, auto-commit, history queries
    ollama/                — Ollama embed endpoint client
    vault/                 — the four vault operations (read/write/edit/search)
    facts/                 — fact extraction, storage, incremental consolidation
    server/                — TCP/HTTP server, command registry, routing
    commands/              — command definitions, init() registration
    keys/                  — constants (command names, event channels, config keys)
  extensions/
    pi/
      pigo.ts              — pi harness extension (registers tools with AI)
  docs/                    — you are here
  vault/                   — default vault location (gitignored, configurable)
```

**Naming conventions:**
- One exported type per file
- Files under 100 lines where possible
- Package names are singular (`vault`, not `vaults`)
- No `internal/utils/` — if it doesn't belong to a domain, it doesn't exist

---

## Dependency Graph

```
cmd/pigo/main.go
  ├── internal/config
  ├── internal/server
  │     └── internal/commands
  │           ├── internal/vault (layer 2)
  │           │     ├── internal/db      (layer 1)
  │           │     ├── internal/git     (layer 1)
  │           │     └── internal/ollama  (layer 1)
  │           └── internal/facts (layer 2)
  │                 ├── internal/db      (layer 1)
  │                 └── internal/ollama  (layer 1)
  └── internal/db (for schema init)
```

Layer 1 packages never import each other. Layer 2 packages may share layer 1 dependencies but don't import each other. Layer 3 (commands) imports layer 2 only.

---

## The Vault Stack

```
Markdown files        — truth layer. Human-readable, editable anywhere.
SQLite + sqlite-vec   — search index. Fuzzy on title/tags, semantic on chunks.
go-git                — history. Every write is a commit. Full versioning.
Ollama                — embeddings. Local, no external API dependency.
```

**Why SQLite, not Postgres:** Single file, ships with the binary, zero config. sqlite-vec gives vector search without Qdrant. The whole system is portable — copy the vault dir and the .db file and you're done.

**Why go-git, not shelling out:** Pure Go, no binary dependency. The gateway binary is self-contained. Push to remote (Gitea, GitHub) is manual or scheduled.

**Why Ollama, not OpenAI:** Local-first. No API key, no network dependency for core operations. Configurable — users can point at any embedding endpoint.

---

## Data Model

### Notes (the vault)

```sql
notes (
  id          TEXT PRIMARY KEY,   -- slug from filename
  file_path   TEXT NOT NULL,
  title       TEXT NOT NULL,
  tags        TEXT NOT NULL,      -- JSON array
  created_at  DATETIME,
  updated_at  DATETIME
)

chunks (
  id          TEXT PRIMARY KEY,   -- uuid
  note_id     TEXT NOT NULL,
  heading     TEXT,               -- heading text, null for intro chunk
  anchor      TEXT,               -- #heading-slug for deep linking
  content     TEXT NOT NULL,
  embedding   BLOB                -- sqlite-vec float32 vector
)
```

### Facts (extracted knowledge)

```sql
facts (
  id          TEXT PRIMARY KEY,
  source_id   TEXT NOT NULL,      -- note id this came from
  content     TEXT NOT NULL,      -- the extracted claim
  topic       TEXT,               -- e.g. "architecture", "people", "decisions"
  importance  INTEGER,            -- 1-10, AI-assigned
  entities    TEXT,               -- JSON array of mentioned entities
  created_at  DATETIME
)
```

Notes are raw material. Facts are refined knowledge extracted by the consolidation pass.

---

## Command Registry

Commands register themselves via Go's `init()` pattern. No manual wiring in main.

```go
// internal/commands/vault_read.go
func init() {
    registry.Register("vault.read", &VaultReadCommand{})
}
```

The server dispatches by command name. The CLI dispatches by subcommand. Both hit the same command implementations.

**Wire format (JSON-RPC inspired):**

```json
// Request
{ "command": "vault.search", "args": { "q": "auth patterns" } }

// Response
{ "success": true, "data": [...] }

// Error
{ "success": false, "error": "note not found: xyz" }
```

**Method catalog:** `system.methods` returns all registered commands with descriptions. AI-discoverable — no out-of-band documentation needed.

---

## The Write Pipeline

When `vault.write` or `vault.edit` is called:

```
1. Save markdown file to disk (with YAML frontmatter)
2. Parse frontmatter + split on headings → chunks
3. For each chunk: call Ollama embed → float32 vector
4. Upsert note row in SQLite
5. Delete old chunk rows for this note
6. Insert new chunk rows with embeddings
7. git add + git commit
8. (async) Search for related notes → update relates_to in frontmatter → recommit
```

Steps 3-6 can be batched. Step 7 is synchronous — commit before returning success.

---

## Server Model

The server runs as a daemon (background process). Two modes:

- **TCP** — persistent connections for the pi extension (pipe model)
- **HTTP** — RESTful for everything else (CLI can use either)

The CLI can operate in two modes:
1. **Direct** — no server needed. CLI loads config, opens the DB, runs the command.
2. **Client** — CLI sends command to running server via TCP/HTTP.

Direct mode means the vault is usable immediately — no daemon required. The server adds: persistent state, background jobs, extension connectivity.

---

## Pi Extension Model

The extension (`extensions/pi/pigo.ts`) runs inside a pi session:

```typescript
export default function (pi: ExtensionAPI) {
  // Register vault tools natively with the AI
  pi.registerTool({
    name: "vault_search",
    label: "Vault Search",
    description: "Search the pigo knowledge vault",
    parameters: Type.Object({ q: Type.String() }),
    async execute(toolCallId, params, signal, onUpdate, ctx) {
      const result = await pi.exec("pigo", ["vault", "search", params.q]);
      return { type: "text", text: result.stdout };
    }
  });

  // ... vault_read, vault_write, vault_edit
}
```

The AI sees native tools, not bash commands. Under the hood, the extension shells out to the `pigo` binary (or connects to the server if running).

---

## Configuration

TOML config file. Location: `~/.pigo/config.toml`.

```toml
[vault]
path = "~/.pigo/vault"        # where notes live on disk

[db]
path = "~/.pigo/pigo.db"      # SQLite database location

[ollama]
endpoint = "http://localhost:11434"
model = "nomic-embed-text"    # embedding model

[server]
host = "127.0.0.1"
port = 14159
pipe_port = 14160

[git]
auto_commit = true
remote = ""                   # empty = no auto-push
```

Everything has sensible defaults. Zero config gets you a working vault.

---

## C# → Go Mental Map

For readers coming from the Gaia.6 C# codebase:

| Gaia.6 C# | pigo Go |
|---|---|
| `Lib/{Name}/` — one external integration | `internal/{name}/` — same boundary |
| `Bloc/{Feature}/` — domain logic | `internal/{feature}/` — same concept |
| `{Module}Services.cs` — DI registration | `init()` + explicit wiring in `main.go` |
| `IService` interface + implementation | Go interface defined where consumed |
| Primary constructor DI | Struct with explicit field injection |
| `BaseDataService` inheritance | Composition — embed a `store` field |
| `IDbContextExtensions` auto-discovery | `init()` registration per package |
| Constants in `Shared/Events/` | `internal/keys/` package |
| Redis pub/sub events | Go channels (in-process) or simple event bus |
| PostgreSQL + Qdrant + Neo4j | SQLite + sqlite-vec (one file, same power for personal scale) |
| HostedService background workers | Goroutines with context cancellation |
| SignalR real-time | TCP pipe / WebSocket |

---

## Design Principles

1. **One binary.** No Docker, no Postgres, no Redis. `go build` and run.
2. **Local-first.** Works offline. Ollama runs locally. No cloud dependency.
3. **Human-readable storage.** Markdown files you can open in any editor.
4. **AI as participant.** The AI builds the knowledge graph, not just queries it.
5. **Layer discipline.** Three layers, no skipping, no exceptions.
6. **Files under 100 lines.** If it's longer, split it.
7. **No magic strings.** Constants package for all keys.
8. **Clear Go.** No clever idioms. Code that a C# developer can read and learn from.
