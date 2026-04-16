# Idea: Agents

Distilled from Gaia.6's Agent feature. This is the concept — not the C# implementation.

---

## The Concept

An agent is an autonomous LLM-powered entity that can receive instructions, maintain conversation context, and work independently. Agents are heavier than a single tool call but lighter than a full session — they're for tasks that need multi-turn reasoning or domain specialization.

---

## When to Use Agents vs Tools

| Need | Use |
|------|-----|
| Simple data operation | Tool (vault.read, vault.search) |
| Structured extraction | Specialized agent (FactAgent) |
| Multi-turn reasoning | General agent |
| Background analysis | Agent with job scheduling |

Tools are stateless and fast. Agents carry context and reason.

---

## Agent Lifecycle

```
Create (name it, give it a system prompt)
  → Active (ready for messages)
  → Send message → Processing → Response
  → Send another message (context preserved)
  → Dispose (clean up)
```

Agents live in memory. They don't persist across restarts (their *results* might, but the agent itself is ephemeral).

---

## Specialized Agents in Gaia.6

Gaia.6 used purpose-built agents for specific tasks:

- **FactAgent** — extracts structured facts from text (JSON output)
- **MemoryAgent** — consolidates and reflects on accumulated memories
- **EmotionalCascadeAgent** — analyzes text through multiple emotional perspectives
- **RelationshipAgent** — discovers connections between entities
- **SceneAgent** — synthesizes system-wide narrative from events
- **NarrativeAgent** — creates coherent stories from activity streams

Each agent has a specialized system prompt that shapes its reasoning.

---

## How This Maps to pigo

pigo's agents would be simpler — no emotional analysis, no narrative synthesis. The relevant patterns:

1. **FactAgent** — already planned for Phase 4 (fact extraction). Takes note content, returns structured JSON facts.

2. **ConsolidationAgent** — could run nightly to review new facts, identify clusters, update relationships across the vault.

3. **SummaryAgent** — could generate periodic summaries ("what happened this week") stored as vault notes.

4. **Custom agents** — users could create agents with their own system prompts for domain-specific analysis.

---

## Implementation Sketch

An agent in pigo is:
- A system prompt (stored as text or as a vault note)
- A conversation history (in-memory slice of messages)
- A connection to the LLM (Ollama or configurable)
- An execute function: send message → get response → append to history

```
Agent{
  Name:     "fact-extractor"
  System:   "You are a fact extraction agent. Given text, identify..."
  History:  []Message
  Model:    "llama3"
}
```

No need for agent hubs, coordination services, or Redis events. Just a struct with a conversation loop. The complexity comes from what the agent *does*, not from the infrastructure around it.

---

## Open Questions for pigo

- **Agent storage:** Should agents be stored in the DB (survive restarts) or purely in-memory?
- **Agent tools:** Should agents be able to call vault tools (read/write/search) during their reasoning? This makes them more powerful but adds complexity.
- **Agent exposure:** Should agents be available as commands (`pigo agent create`, `pigo agent message`)? Or only used internally by the job system?

---

## Why This Matters

Agents are the mechanism that turns pigo from a storage system into a thinking system. The vault stores knowledge. Tools query it. But agents *reason* about it — extracting facts, finding patterns, generating summaries. They're the bridge between raw data and structured understanding.
