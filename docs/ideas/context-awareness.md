# Idea: Context Awareness

What separates a search box from an assistant.

---

## The Problem

Every new AI session starts cold. The AI doesn't know what project you're in, what you've been working on, what your preferences are, or what happened yesterday. You re-explain everything, every time.

pigo can fix this by making context automatic.

---

## Three Layers of Context

### 1. User Context (static, rarely changes)

`~/.pigo/preferences.toml`:
```toml
name = "Scot"
role = "Head of Development"
expertise = ["C#", "cybersecurity", "ICAM"]
learning = ["Go"]
timezone = "America/Chicago"
style = "direct, no fluff, no summaries"
```

Injected into every session's system prompt. The AI always knows who it's talking to.

### 2. Project Context (changes per directory)

`pigo context project` scans the current directory:
- Detect project type (Go, Node, C#, Python) from manifest files
- Read README.md, CLAUDE.md for project description
- List recent git activity (last 5 commits)
- Return structured summary

The pi extension injects this on session start. The AI knows "I'm in a Go project called pigo" without being told.

### 3. Session Context (changes per conversation)

Track within a session:
- What notes were read/written/searched
- What facts were referenced
- What topics came up

`pigo context session` returns the activity log. Useful for:
- "What have we been working on?" — instant answer without re-reading
- Session-end summary — offer to save highlights as a vault note

---

## System Prompt Injection

The pi extension hooks `before_agent_start` and injects:

```
## About the User
{contents of preferences.toml}

## Current Project
{output of pigo context project}

## Recent Vault Activity
{last 5 notes written or searched}

## Available Tools
pigo is running. You have access to a persistent knowledge vault.
Use vault_search to find prior knowledge. Use vault_write to save new knowledge.
```

This happens automatically. The user doesn't configure anything beyond initial preferences.

---

## Conversation Memory

On session end (or when the user says `/remember`):
1. Summarize key decisions, learnings, and actions from the conversation
2. Offer to save as a vault note with appropriate tags
3. If saved, it's searchable in future sessions

This closes the loop: conversations produce knowledge, knowledge informs future conversations.

---

## Why This Matters

The AI that knows your name, your project, and what you worked on yesterday is dramatically more useful than one that starts fresh every time. Context awareness is what makes pigo feel like a persistent assistant rather than a stateless tool.
