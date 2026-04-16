# The Vault

The vault is the knowledge layer. It's the first thing we're building and arguably the most important capability the gateway has.

---

## What It Is

A directory of markdown files. Each file is a note. Notes have YAML frontmatter for metadata (title, tags, created, updated, relationships). The body is plain markdown.

The vault looks like this on disk:

```
vault/
  decisions/
    2026-04-15-chose-go-over-csharp.md
    2026-03-01-dropped-neo4j.md
  projects/
    pigo-new.md
    90meter-auth-rewrite.md
  people/
    scot.md
  observations/
    2026-04-16-markdown-as-knowledge-graph.md
```

Human-readable. Editable in any editor. Backed by git. Indexed by SQLite.

---

## Schema

**Frontmatter (YAML):**
```yaml
---
title: Why we chose Go
tags: [architecture, go, decisions]
created: 2026-04-15T10:30:00Z
updated: 2026-04-15T10:30:00Z
relates_to: [pigo-new, dropped-neo4j]
---
```

`relates_to` is AI-generated on write. The AI decides what this note connects to based on content similarity and existing notes in the vault. No manual linking.

**SQLite tables:**

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

---

## The Four Tools

### `vault.read`
Read a note by title, slug, or path. Returns frontmatter + body.

```json
{ "command": "vault.read", "args": { "id": "why-we-chose-go" } }
```

### `vault.write`
Create a new note. Triggers: frontmatter generation, chunking, embedding, git commit.

```json
{
  "command": "vault.write",
  "args": {
    "title": "Why we chose Go",
    "tags": ["architecture", "go"],
    "body": "..."
  }
}
```

### `vault.edit`
Update an existing note (body or tags). Re-indexes only the changed file. Git commit.

```json
{
  "command": "vault.edit",
  "args": {
    "id": "why-we-chose-go",
    "body": "... updated content ..."
  }
}
```

### `vault.search`
Search by meaning, not just keywords. Runs two passes and merges:
1. Fuzzy match on title and tags (fast, no embedding needed)
2. Semantic similarity on chunk embeddings via sqlite-vec

Returns a ranked list: note title, matching heading, relevance score, path + anchor.

```json
{ "command": "vault.search", "args": { "q": "why we picked go over csharp" } }
```

---

## The Write Pipeline

When `vault.write` or `vault.edit` is called:

```
1. Save markdown file to disk
2. Parse frontmatter + split on headings → chunks
3. For each chunk: call Ollama embed → float32 vector
4. Upsert note row in SQLite
5. Delete old chunk rows for this note
6. Insert new chunk rows with embeddings
7. git add <file> && git commit -m "vault: <title>"
```

Steps 3–6 can be batched. Step 7 is synchronous — the commit happens before the tool returns success.

---

## Relationships (AI-generated)

On write, after the note is saved, the gateway runs a search for notes similar to the new content. The top N results become the `relates_to` list in the frontmatter. The file is updated, re-committed.

This means the graph builds itself. New notes find their neighbors automatically. Over time, clusters emerge from actual content similarity — not from whatever the human remembered to link.

---

## Memory vs Fact (from Gaia.6)

Gaia.6 drew a clean line between two things we should honor:

**Notes (memory)** — quick captures and rich observations. Atomic, fast to write. "Alex mentioned RL." "Decided to drop Neo4j." "The vault idea clicked." Raw material.

**Facts** — extracted wisdom. Structured, typed, tagged with importance. Not a narrative, not a raw note — a discrete, durable claim about the world.

In our system, notes live in the vault. Facts are extracted *from* notes by a background consolidation pass and stored separately in SQLite as structured rows with topic, importance, and entity links. The vault is the raw layer; the fact table is the refined layer.

```sql
facts (
  id          TEXT PRIMARY KEY,
  source_id   TEXT NOT NULL,   -- note id this came from
  content     TEXT NOT NULL,   -- the extracted claim
  topic       TEXT,            -- e.g. "architecture", "people", "decisions"
  importance  INTEGER,         -- 1-10, AI-assigned
  entities    TEXT,            -- JSON array of mentioned entities
  created_at  DATETIME
)
```

---

## Nightly Consolidation

On a schedule (nightly or configurable), the consolidation agent:

1. Finds notes written since last run
2. For each note, runs an LLM prompt: "Extract discrete facts from this note. For each fact: content, topic, importance 1-10, entities mentioned."
3. Stores extracted facts in the `facts` table
4. Updates `relates_to` in frontmatter based on fact overlap with other notes
5. Commits the updated frontmatter

This is the mechanic that makes memory *compound* rather than just accumulate. Raw notes become structured knowledge over time without requiring the author to do anything beyond writing.

Gaia.6 called this MemoryAgent. We'll call it the consolidation pass. Same idea.

---

## Git Integration

Uses `go-git` — pure Go, no binary dependency, ships inside the gateway binary.

On `vault.write` / `vault.edit`:
- Auto-commit with message: `vault: <title> [write|edit]`

History queries (future):
- `vault.history <id>` — commits touching this file
- `vault.diff <id> <from> <to>` — content diff between versions
- `vault.restore <id> <commit>` — revert a note to a prior version

The remote is Gitea on tower. Push is not automatic — triggered manually or on a schedule.
