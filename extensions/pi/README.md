# pigo Pi Extension

Gives the AI native vault tools inside a [pi](https://github.com/badlogic/pi-mono) coding session.

## Install

Symlink or copy the extension into your pi extensions directory:

```bash
# Symlink (recommended — stays in sync with repo)
ln -s /path/to/pigo/extensions/pi/pigo.ts ~/.pi/extensions/pigo.ts

# Or copy
cp /path/to/pigo/extensions/pi/pigo.ts ~/.pi/extensions/pigo.ts
```

## Requirements

- **pigo binary** on your PATH (or the daemon running)
- **Ollama** running locally for embeddings (optional — fuzzy search works without it)

## How It Works

The extension connects to the pigo daemon via TCP pipe on port 14160. If the daemon isn't running, it falls back to calling the `pigo` CLI directly.

### Tools Registered

| Tool | Description |
|------|-------------|
| `vault_search` | Search the vault by meaning, keywords, or tags |
| `vault_read` | Read a note by ID |
| `vault_write` | Create a new note (indexed, embedded, committed) |
| `vault_edit` | Update an existing note |
| `vault_list` | List all notes |
| `pigo_command` | Execute any registered pigo command |

### System Prompt

If `~/.pigo/system.md` exists, its contents are injected into the AI's system prompt on every session. Use this to teach the AI about your vault, preferences, or workflow.

### Slash Commands

- `/vault <query>` — Quick vault search
