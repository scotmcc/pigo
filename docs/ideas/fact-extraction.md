# Idea: Fact Extraction

Distilled from Gaia.6's Fact feature. This is the concept — not the C# implementation.

---

## The Concept

Notes are raw material. Facts are refined knowledge. An LLM reads notes and extracts discrete, atomic claims — structured, typed, scored for importance. Over time, the vault of raw notes produces a layer of distilled wisdom that's searchable, filterable, and connected.

This is what makes memory *compound* rather than just accumulate.

---

## What a Fact Looks Like

Given a note:
> "Talked to Alex about machine learning. She was excited about reinforcement learning. We discussed how it relates to consciousness."

The extraction produces:

```json
[
  {
    "content": "Alex is interested in reinforcement learning",
    "topic": "interests",
    "importance": 7,
    "entities": ["Alex", "reinforcement learning"]
  },
  {
    "content": "Reinforcement learning may relate to consciousness",
    "topic": "philosophy",
    "importance": 8,
    "entities": ["reinforcement learning", "consciousness"]
  }
]
```

Each fact is:
- **Atomic** — one claim, not a paragraph
- **Topical** — categorized (open-ended, not a fixed list)
- **Scored** — importance 1-10 (AI-assigned)
- **Entity-linked** — people, concepts, places mentioned

---

## The Extraction Pipeline

```
Note content
  → LLM with extraction prompt (system prompt + note body)
  → JSON array of facts
  → Parse and validate
  → Store in facts table with source linkage
  → Index for search (semantic via embeddings, or topic/importance filters)
```

The LLM prompt instructs:
- Identify discrete facts (not summaries, not opinions)
- Assign topics naturally (user's own categories emerge)
- Rate importance (1-3 trivial, 4-6 interesting, 7-8 important, 9-10 critical)
- Extract mentioned entities
- Return valid JSON

---

## Incremental Extraction

Track when each note was last extracted (`last_extraction_at`). On each consolidation run:

1. Find notes where `updated_at > last_extraction_at` (or `last_extraction_at` is null)
2. Extract facts only from those notes
3. Update `last_extraction_at`

This means nightly runs process only new/changed notes. A full rebuild is available for prompt changes or model upgrades but shouldn't be needed regularly.

**Failure handling:** If extraction fails for a note, don't update the timestamp — it'll be retried next run. Log the error, keep going.

---

## Search Patterns

Facts support multiple query patterns:

| Query | How |
|-------|-----|
| Semantic search | Embed the query, cosine similarity against fact embeddings |
| By topic | Filter on topic field |
| By importance | Filter on importance >= threshold |
| Recent | Order by created_at desc |
| By entity | Filter on entities JSON array |
| By source | Find all facts extracted from a specific note |

---

## Relationship to the Vault

```
Vault (notes)          →  raw material, human-readable, versioned
  ↓ extraction
Facts table            →  refined knowledge, structured, searchable
```

Notes are the truth layer. Facts are the intelligence layer. The extraction prompt can evolve without touching the notes — just re-extract with a better prompt.

---

## Open Design Questions for pigo

- **Embedding facts:** Should facts get their own embeddings in sqlite-vec, or rely on the note chunks they came from? (Separate embeddings = better fact-level search, more storage)
- **Entity tracking:** Gaia.6 used Neo4j for entity relationships. In pigo, entities could be a simple SQLite table with fact-entity links. Is that enough?
- **Deduplication:** What happens when two notes produce the same fact? Deduplicate by content similarity? Keep both with different sources?
- **Fact lifecycle:** Can facts be manually corrected or deleted? Or are they always derived from notes?

---

## Why This Matters

Without fact extraction, the vault is a pile of notes. With it, the vault becomes a knowledge base. You can ask "what do I know about X?" and get structured, scored, sourced answers — not just "here's a note that might be relevant."

The nightly consolidation pass is what turns a note-taking app into a second brain.
