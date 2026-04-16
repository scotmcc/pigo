---
name: pigo
description: Search and manage the pigo knowledge vault
---

# pigo — Knowledge Vault

You have access to a persistent knowledge vault via the `pigo` CLI. The vault stores markdown notes that are indexed for search, embedded for semantic matching, and versioned with git.

## First: check if you know the user

Run `pigo soul` at the start of a session. If a soul file exists, read it — it tells you who the user is, how they work, and what they care about. If no soul exists, consider asking the user about themselves and creating one:

```bash
pigo vault write --title "Soul" --tags "system,identity" --body "## User\n- Name: ..."
```

## When to use

- **Searching for prior knowledge:** When the user asks about something that might be in the vault, or when you need context from past work, search first.
- **Saving new knowledge:** When you learn something worth remembering across sessions — decisions, patterns, architecture notes, useful findings — save it.
- **Reviewing what's known:** When starting work on a topic, check what's already in the vault.

## Commands

### Search the vault
```bash
pigo vault search "query terms"
```
Returns ranked results matching by title, tags, and content. Use `--json` for structured output.

### Read a note
```bash
pigo vault read <note-id>
```
The note ID is the slug (e.g., `why-we-chose-go`). Returns full markdown with frontmatter.

### Create a note
```bash
pigo vault write --title "Title Here" --tags "tag1,tag2" --body "Markdown content here"
```
Creates a new note, indexes it, embeds it for semantic search, and commits to git. Use `[[slug]]` syntax in the body to link to other notes.

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

### Check tags
```bash
pigo vault tags
```
Shows all tags with note counts. Check this before inventing new tags.

### View connections
```bash
pigo vault links <note-id>
```
Shows what a note relates to, links to, and what links back to it.

### Import a web page
```bash
pigo vault import <url> --tags "tag1,tag2"
```
Fetches a URL, converts HTML to markdown, and saves it as a vault note.

### Search the web
```bash
pigo web search "query terms"
```
Searches via SearXNG (if configured). Use `vault import` to save useful results.

## Guidelines

- Search before writing — avoid duplicating knowledge that already exists.
- Use descriptive titles that will be findable later.
- Tag consistently — run `pigo vault tags` before inventing new tags.
- Link related notes with `[[slug]]` syntax in the body.
- Keep notes atomic — one topic per note, not a dump of everything.
- When you save a note, tell the user what you saved and why.
- Use `--json` when you need to parse the output programmatically.
