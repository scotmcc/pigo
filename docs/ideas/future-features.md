# Idea: Future Features

Concepts from Gaia.6 and other sources that could become pigo features. None of these are planned — they're ideas to revisit when the core (vault + facts + jobs) is solid.

---

## Speech / TTS

**From Gaia.6:** Kokoro TTS integration. The system can speak — choose a voice, synthesize audio, play it. Voice selection driven by context (emotional state, content type).

**In pigo:** A `tts.speak` command that calls a TTS service (Kokoro, or any configurable endpoint). The pi extension could play audio in the session. The CLI could output to speakers or save to file.

**Why it's interesting:** Makes the knowledge system tangible. Instead of reading search results, hear them spoken. Useful for accessibility, hands-free operation, or just making the system feel more alive.

**Complexity:** Medium. It's an integration (layer 1) + a thin command (layer 3). No business logic needed.

---

## Scene / Narrative

**From Gaia.6:** A meta-observer that watches all system events and synthesizes "what's happening right now" into a coherent narrative. Uses an LLM agent to produce running commentary.

**In pigo:** A background agent that periodically reviews recent vault activity (new notes, extracted facts, relationship changes) and writes a summary note. "This week: 5 notes about auth architecture, 3 new facts about deployment patterns, relationship cluster forming around security topics."

**Why it's interesting:** The vault generates its own documentation over time. You get a meta-layer of self-awareness about what the knowledge system is learning.

**Complexity:** High. Requires agents + jobs + vault write access. This is a Phase 6+ feature.

---

## Music / Context Integration

**From Gaia.6:** Spotify "now playing" integration. The system knows what music is playing and can weave it into context, history, and emotional analysis.

**In pigo:** A command namespace (`music.*`) that polls a music service and records what's playing. Could be a context signal — "I was listening to X when I wrote this note."

**Why it's interesting:** Context enrichment. Notes become richer when they carry ambient information.

**Complexity:** Low (it's just an API integration) but low value for a distributable tool. This is personal-infrastructure territory.

---

## History / Timeline

**From Gaia.6:** Event-driven history that records significant moments (agent messages, commands, state changes). An agent periodically analyzes the timeline for meaning.

**In pigo:** The vault already has git history. A `history` command could surface activity timelines — what was written, when, what facts were extracted. Git log is the foundation; a richer timeline layer could aggregate across vault + facts + jobs.

**Why it's interesting:** Answers "what happened last Tuesday?" or "show me the evolution of my thinking about X."

**Complexity:** Low for basic git-based history. Higher for cross-system timeline aggregation.

---

## Skills (Stored Prompts)

**From Gaia.6:** Skills are named, stored system prompts. Jobs execute skills. Agents use skills for domain specialization.

**In pigo:** Skills could be vault notes with a special tag (`type: skill`). A skill is just markdown — the system prompt is the note body. Jobs reference skills by slug. This means skills are versioned (git), searchable (vault.search), and human-editable.

**Why it's interesting:** No separate skill storage needed. The vault *is* the skill store. And skills benefit from the same versioning, search, and relationship features as any other note.

**Complexity:** Low. It's a convention on top of the vault, not a separate feature.

---

## Docker / System Integration

**From Gaia.6:** System commands for container management, process control, etc.

**In pigo:** A `system.*` command namespace. `system.exec` runs arbitrary commands (with safety controls). `docker.*` manages containers. This turns pigo into a lightweight system management layer.

**Why it's interesting:** Makes pigo useful beyond knowledge management. The AI can manage services, check logs, restart containers — all through the same command interface.

**Complexity:** Medium, and needs careful security design. Shell execution is powerful and dangerous.

---

## Web UI

**In pigo:** A browser-based vault explorer. Browse notes, view relationships, search, see fact graphs. Served by the pigo HTTP server.

**Why it's interesting:** Not everyone wants a CLI. A web UI makes the vault accessible to non-technical users and provides visualization that terminals can't.

**Complexity:** High. Frontend development is a different skill set. Consider whether this is worth building vs. pointing people at the markdown files directly. The vault is human-readable by design.

---

## Priority Thinking

If you're looking at this list and wondering what to build next after the core:

1. **Skills as vault notes** — almost free, just a convention
2. **Basic history** — git log is already there, just surface it
3. **Speech/TTS** — clean integration, high impact for the right user
4. **Scene/narrative** — requires agents and jobs, but produces the most interesting emergent behavior

Everything else is nice-to-have.
