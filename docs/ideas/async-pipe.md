# Idea: Async Commands + Persistent Pipe

The proven architecture from old pigo that enables real-time AI interaction.

---

## The Problem

HTTP request/response is fine for simple queries. But some operations take time — embedding 50 notes, consolidating facts, importing a large webpage. The AI needs to know what's happening, not just wait for a final result.

---

## The Solution: Persistent Pipe + GUID Routing

### Persistent Pipe

Instead of a new connection per request, the pi extension opens one TCP connection at startup and keeps it alive for the entire session.

```
Extension starts → TCP connect to pigo daemon
Extension sends  → {"type": "register_pipe"}
Daemon responds  → {"status": "ok", "message": "pipe registered"}
Connection stays open — all async responses flow through it
```

Newline-delimited JSON. One line = one message. Simple to parse, simple to debug.

### GUID Routing

Every command gets a unique GUID (generated client-side). The daemon includes the GUID in every response. The extension matches responses to the tool call that triggered them.

This means multiple commands can be in flight simultaneously on one connection. No blocking, no ordering issues.

### Async Command Flow

```
1. AI calls vault_write tool
2. Extension generates GUID, sends {guid, command, args} through pipe
3. Daemon sees async flag, immediately returns {guid, status: "accepted"}
4. Extension returns "accepted" to AI
5. Daemon processes in background goroutine
6. Daemon sends {guid, status: "update", message: "Embedding chunk 3/7..."} through pipe
7. Extension receives update, sends as steer message to AI
8. Daemon sends {guid, status: "done", data: {id: "...", path: "..."}}
9. Extension receives done, steer final result to AI
```

### SendFunc Pattern

Commands don't know about TCP, pipes, or GUIDs. They receive a callback:

```go
type SendFunc func(status, message string, data any) error

func (c *consolidateCmd) Execute(args map[string]any, send SendFunc) (any, error) {
    notes := getUnprocessedNotes()
    for i, note := range notes {
        send("update", fmt.Sprintf("Processing %d/%d: %s", i+1, len(notes), note.Title), nil)
        extractFacts(note)
    }
    return result, nil
}
```

The command just calls `send()`. Whether the response goes to an HTTP connection, a TCP pipe, or stdout doesn't matter. Clean separation.

---

## Sync vs Async

A flag on the command registration:

```go
commands.Register("vault.read", &vaultReadCmd{}, commands.Info{Async: false})
commands.Register("facts.consolidate", &consolidateCmd{}, commands.Info{Async: true})
```

Sync commands: response sent on the same connection, blocking.
Async commands: immediate ack, progress via pipe, final result via pipe.

The caller (HTTP handler or pipe handler) checks the flag and behaves accordingly.

---

## Why This Matters

Without async + pipe:
- AI calls a tool → waits → gets result (or timeout)
- No progress visibility
- Long operations feel broken

With async + pipe:
- AI calls a tool → gets ack → sees progress → gets result
- "Processing 3/7..." steer messages keep the AI informed
- Long operations feel alive

This is what makes pigo feel responsive and real-time, not like a batch processor.
