# Idea: Job System

Distilled from Gaia.6's Job and Cron features. This is the concept — not the C# implementation.

---

## The Concept

A job system gives pigo a heartbeat. Without it, the server only responds to requests. With it, the server *acts* — running consolidation passes, refreshing relationships, monitoring changes, executing scheduled work.

---

## Two Tiers of Jobs

### System Jobs
Built-in, always running, not user-manageable. They're the infrastructure:

- **Nightly consolidation** — extract facts from new/changed notes
- **Relationship refresh** — re-evaluate `relates_to` across the vault
- **Health check** — verify Ollama is reachable, DB is healthy

System jobs are registered at startup. They can't be paused or deleted — they're part of the binary.

### Dynamic Jobs
User-created, stored in the database, fully manageable:

- Created via CLI or API: `pigo jobs create --name "weekly-report" --cron "0 9 * * 1" --command "vault.search" --args '{"q":"decisions this week"}'`
- Each job wraps a command (from the command registry) with arguments
- Can be paused, resumed, canceled
- State survives restarts (stored in SQLite)

---

## Job Lifecycle

```
Created (stored in DB, registered with scheduler)
  → Running (executes on cron schedule)
    → Success → next execution scheduled
    → Failure → error logged, continues on schedule
  → Paused (unregistered from scheduler, stays in DB)
    → Resumed (re-registered)
  → Canceled (removed from DB)
```

---

## Cron Scheduling

Standard cron expressions:
- `0 2 * * *` — daily at 2 AM
- `*/5 * * * *` — every 5 minutes
- `0 * * * *` — every hour

The scheduler is a goroutine that checks registered jobs every 60 seconds against the current time. When a cron expression matches, the job executes.

---

## Skills as Job Payloads (Future Idea)

In Gaia.6, jobs execute "Skills" — stored system prompts that guide LLM behavior. The concept:

- A Skill is a named, stored prompt template: `{ name: "summarize-week", prompt: "Review notes from the past 7 days and produce a summary..." }`
- A Job references a Skill by name and runs it on schedule
- The Skill prompt can be refined without touching the job definition

This separates *when to run* (the job) from *what to do* (the skill). pigo could adopt this pattern — Skills stored as notes in the vault, referenced by jobs.

---

## Execution Model

```
Scheduler tick (every 60s)
  → Check all registered jobs against current time
  → For matching jobs:
    → Look up command in registry
    → Execute with stored args
    → Record result (success/failure, duration, output)
    → Update next_run timestamp
```

Jobs execute sequentially by default (no parallel execution of the same job). Different jobs can run concurrently.

**Error handling:** Log the error, record it on the job record (`last_error`), continue scheduling. Don't crash the server. Don't retry immediately — wait for next scheduled run.

---

## Data Model

```sql
jobs (
  id          TEXT PRIMARY KEY,
  name        TEXT NOT NULL UNIQUE,
  cron        TEXT NOT NULL,       -- cron expression
  command     TEXT NOT NULL,       -- command registry name
  args        TEXT,                -- JSON args for the command
  status      TEXT NOT NULL,       -- scheduled, paused, canceled
  last_run    DATETIME,
  next_run    DATETIME,
  last_error  TEXT,
  created_at  DATETIME
)
```

---

## Why This Matters

The job system is what makes pigo *alive* between sessions. Without it, knowledge only grows when someone actively writes notes. With it, the system maintains itself — consolidating facts, refreshing relationships, running health checks — whether anyone is talking to it or not.
