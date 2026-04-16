# Idea: Web Import

Distilled from old pigo's `web` command. Upgraded for the vault.

---

## The Concept

The AI encounters a useful URL during a conversation. Instead of just reading it and forgetting, it saves the content as a vault note — indexed, embedded, versioned, searchable forever.

```
pigo vault import https://go.dev/doc/effective_go
```

Creates a note with:
- Title extracted from the page's `<title>` or `<h1>`
- Body converted from HTML to clean markdown
- Frontmatter with `source_url`, `imported_at`, and auto-generated tags
- Embedded, indexed, committed — just like any other note

---

## HTML → Markdown

Old pigo used regex-based conversion. No DOM parser, no heavy dependency. The approach:

1. Strip `<script>`, `<style>`, `<nav>`, `<footer>` blocks
2. Convert `<h1>`-`<h6>` → `#` headings
3. Convert `<p>` → paragraphs with blank lines
4. Convert `<a href="...">text</a>` → `[text](href)`
5. Convert `<code>` → backticks, `<pre>` → code blocks
6. Convert `<li>` → `- ` list items
7. Strip remaining tags
8. Collapse whitespace

Not perfect, but good enough for knowledge capture. If we need better fidelity later, swap in a real parser.

---

## Frontmatter

```yaml
---
title: Effective Go
tags: [go, imported, go.dev]
source_url: https://go.dev/doc/effective_go
imported_at: 2026-04-16T10:30:00Z
created: 2026-04-16T10:30:00Z
updated: 2026-04-16T10:30:00Z
---
```

The `source_url` field is unique to imported notes. Useful for deduplication (don't import the same URL twice) and attribution.

---

## Why This Matters

The AI's context window is temporary. The vault is permanent. Web import bridges that gap — anything the AI reads from the web can become part of the persistent knowledge base. Over time, the vault accumulates reference material that makes the AI more useful in every future session.
