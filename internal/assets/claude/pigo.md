---
name: pigo
description: Search and manage the pigo knowledge vault
---

# pigo — Knowledge Vault

You have access to a persistent knowledge vault via the `pigo` CLI. The vault stores markdown notes that are indexed for search, embedded for semantic matching, and versioned with git.

## When to use

- **Searching for prior knowledge:** When the user asks about something that might be in the vault, or when you need context from past work, search first.
- **Saving new knowledge:** When you learn something worth remembering across sessions — decisions, patterns, architecture notes, useful findings — save it.
- **Reviewing what's known:** When starting work on a topic, check what's already in the vault.

## Commands

### Search the vault
```bash
pigo vault search "query terms"
```
Returns ranked results matching by title, tags, and content. Use this liberally — it's fast.

### Read a note
```bash
pigo vault read <note-id>
```
The note ID is the slug (e.g., `why-we-chose-go`). Returns full markdown with frontmatter.

### Create a note
```bash
pigo vault write --title "Title Here" --tags "tag1,tag2" --body "Markdown content here"
```
Creates a new note, indexes it, embeds it for semantic search, and commits to git.

For longer content, pipe from stdin:
```bash
echo "Long content here..." | pigo vault write --title "Title" --tags "tag1" --stdin
```

### Edit a note
```bash
pigo vault edit <note-id> --body "Updated content" --tags "new,tags"
```
Updates the note, re-indexes, re-embeds, and recommits.

### List all notes
```bash
pigo vault list
```
Shows all notes with dates and tags.

## Guidelines

- Search before writing — avoid duplicating knowledge that already exists.
- Use descriptive titles that will be findable later.
- Tag consistently — look at existing tags with `pigo vault list` before inventing new ones.
- Keep notes atomic — one topic per note, not a dump of everything.
- When you save a note, tell the user what you saved and why.
