# Idea: Architectural Patterns

Distilled from Gaia.6's four architectural patterns. Translated from C#/.NET thinking into Go thinking for pigo.

---

## The Four Patterns

Gaia.6 discovered four distinct ways to structure features. Not all apply to pigo today, but the taxonomy is useful for deciding how new features should be built.

---

### 1. Data-Driven

**What:** Feature owns persistent, searchable data. Full CRUD + indexing + events.

**Gaia.6:** Entity → Model → DataService → BusinessService (Memory, Journal, Goal)

**pigo equivalent:** SQLite table + Go service struct + command handlers

**When to use:** The feature needs to store data that survives restarts and is queryable. The vault is data-driven. Facts are data-driven.

**Structure in Go:**
```
internal/{feature}/
  store.go      — SQLite operations (layer 1)
  service.go    — business logic, orchestration (layer 2)
  // commands registered in internal/commands/ (layer 3)
```

---

### 2. Reactive

**What:** Feature reacts to events from other features. Stateless computation triggered by system activity.

**Gaia.6:** HostedService + Redis event subscriptions (Aspect — emotional analysis)

**pigo equivalent:** Goroutine + Go channel subscriptions

**When to use:** Something should happen *in response to* vault writes, fact extraction, or other system events. No data of its own.

**Example for pigo:** A relationship updater that listens for new notes and re-evaluates `relates_to` fields across the vault. It doesn't own data — it modifies vault data in response to events.

**Structure in Go:**
```
internal/{feature}/
  listener.go   — goroutine that subscribes to events
  handler.go    — processing logic (stateless)
```

---

### 3. Meta-Reactive

**What:** Observes *all* system activity and synthesizes a higher-level view. Listens broadly, maintains in-memory state.

**Gaia.6:** SceneHub — subscribes to 10+ event channels, builds a narrative of "what's happening right now"

**pigo equivalent:** Would be a goroutine that watches vault writes, fact extractions, job completions, and produces a system status or summary.

**When to use:** Rarely. This is for system-wide synthesis. pigo probably doesn't need this until much later (if ever).

---

### 4. Pure Orchestration

**What:** Coordinates active requests between services. No persistence, no events — just request/response routing.

**Gaia.6:** CommandRouter, ClientRegistry, AgentCoordinationService

**pigo equivalent:** The command registry + server dispatch. Also agent lifecycle management (create, message, dispose).

**When to use:** Active coordination between components. The server's command dispatch is pure orchestration. Agent management would be pure orchestration.

**Structure in Go:**
```
internal/server/
  router.go     — dispatch commands to handlers
  registry.go   — track registered commands
```

---

## Decision Matrix

| Need | Pattern | pigo Example |
|------|---------|-------------|
| Store and query data | Data-Driven | Vault, Facts, Jobs |
| React to system events | Reactive | Relationship refresh on vault write |
| Synthesize system state | Meta-Reactive | (future) System status/summary |
| Route requests | Pure Orchestration | Command registry, agent lifecycle |

---

## The Key Insight

Most features in pigo are **data-driven** (vault, facts, jobs). The reactive pattern becomes relevant when features need to respond to each other — "when a note is written, update related notes." The orchestration pattern is the server itself.

Don't over-engineer pattern selection. If the feature stores data → data-driven. If it reacts to events → reactive. If it routes requests → orchestration. Most things are data-driven.
