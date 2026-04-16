/**
 * pigo — Pi Extension
 *
 * Gives the AI native tools for the pigo knowledge vault.
 * Connects to the pigo daemon via persistent TCP pipe for real-time streaming.
 * Falls back to CLI execution if the daemon isn't running.
 *
 * On first session with no soul file, guides the AI through a welcome flow
 * to learn about the user. On subsequent sessions, injects the soul into
 * the system prompt so the AI always knows who it's talking to.
 *
 * Install: pigo install (or symlink into .pi/extensions/)
 */

import type { ExtensionAPI } from "@mariozechner/pi-coding-agent";
import {
	DEFAULT_MAX_BYTES,
	DEFAULT_MAX_LINES,
	formatSize,
	truncateHead,
} from "@mariozechner/pi-coding-agent";
import { Type } from "@sinclair/typebox";
import { createConnection, type Socket } from "net";

// -------------------------------------------------------------------
// Config
// -------------------------------------------------------------------

const PIPE_HOST = "localhost";
const PIPE_PORT = 14160;

// -------------------------------------------------------------------
// Pipe state
// -------------------------------------------------------------------

let pipeConn: Socket | null = null;
let pipeBuffer = "";

/** Tracks in-flight async commands by GUID. */
const pendingCommands = new Map<
	string,
	{
		resolve: (value: any) => void;
		reject: (reason: any) => void;
	}
>();

// -------------------------------------------------------------------
// Extension entry point
// -------------------------------------------------------------------

export default function (pi: ExtensionAPI) {
	// Soul-based system prompt injection.
	// Reads the soul from the vault (via CLI fallback since pipe may not be ready).
	// If no soul exists, injects the welcome prompt to start the getting-to-know-you flow.
	pi.on("before_agent_start", async (event) => {
		try {
			const soulPrompt = await getSoulPrompt(pi);
			if (soulPrompt) {
				return { systemPrompt: event.systemPrompt + "\n\n" + soulPrompt };
			}
		} catch {
			// pigo not available — skip soul injection.
		}
		return {};
	});

	// Connect to daemon pipe on session start.
	pi.on("session_start", () => {
		connectPipe(pi);
	});

	// Clean up on shutdown.
	pi.on("session_shutdown", () => {
		if (pipeConn) {
			pipeConn.destroy();
			pipeConn = null;
		}
	});

	// Register tools.
	registerVaultSearch(pi);
	registerVaultRead(pi);
	registerVaultWrite(pi);
	registerVaultEdit(pi);
	registerVaultList(pi);
	registerVaultLinks(pi);
	registerVaultTags(pi);
	registerWebSearch(pi);
	registerPigoCommand(pi);

	// Register slash commands.
	pi.registerCommand("vault", {
		description: "Search the pigo vault",
		handler: async (args, ctx) => {
			const query = args || "recent notes";
			const result = await runCommand(pi, "vault.search", { q: query });
			if (ctx.hasUI) {
				ctx.ui.notify(`Vault: ${formatSearchSummary(result)}`, "info");
			}
		},
	});

	pi.registerCommand("remember", {
		description: "Save a note to the vault from the current conversation",
		handler: async (args, ctx) => {
			if (!args) {
				if (ctx.hasUI) ctx.ui.notify("Usage: /remember <what to remember>", "info");
				return;
			}
			pi.sendUserMessage(
				`Please save the following to the pigo vault as a note with an appropriate title and tags: ${args}`,
				{ deliverAs: "followUp" },
			);
		},
	});
}

// -------------------------------------------------------------------
// Soul system
// -------------------------------------------------------------------

async function getSoulPrompt(pi: ExtensionAPI): Promise<string | null> {
	// Try reading soul via CLI (works whether daemon is running or not).
	try {
		const result = await pi.exec("pigo", ["vault", "read", "soul"]);
		if (result.code === 0 && result.stdout.trim()) {
			// Soul exists — build the prompt.
			// Read the preamble from the soul init command.
			const preambleResult = await pi.exec("pigo", ["soul", "init", "--prompt"]);
			// The preamble is the welcome prompt, but we want the soul preamble.
			// For now, use a simple preamble inline.
			return buildSoulSystemPrompt(result.stdout.trim());
		}
	} catch {
		// pigo not installed or not working.
	}

	// No soul — check if pigo is available at all.
	try {
		const check = await pi.exec("pigo", ["version"]);
		if (check.code === 0) {
			// pigo works but no soul — return the welcome prompt.
			const welcome = await pi.exec("pigo", ["soul", "init", "--prompt"]);
			if (welcome.code === 0) {
				return welcome.stdout;
			}
		}
	} catch {
		// pigo not available.
	}

	return null;
}

function buildSoulSystemPrompt(soulContent: string): string {
	return `# pigo Knowledge Vault

You are connected to pigo, a persistent knowledge vault. You have tools to search, read, write, edit, import, and link notes. The vault persists across sessions — anything you save is available next time.

## Your role

- Search the vault before answering questions that might have prior context
- Save knowledge worth remembering — decisions, patterns, learnings, useful findings
- Use consistent tags (check vault_tags to see what's in use)
- Link related notes with [[slug]] syntax when writing
- The vault grows from use — the more you contribute, the more useful it becomes

## About the user

The following is the user's soul file — their identity, preferences, and context. This was built from conversation and may be updated over time.

---

${soulContent}`;
}

// -------------------------------------------------------------------
// Pipe connection
// -------------------------------------------------------------------

function connectPipe(pi: ExtensionAPI) {
	if (pipeConn) return;

	pipeConn = createConnection({ port: PIPE_PORT, host: PIPE_HOST });
	pipeBuffer = "";

	pipeConn.on("connect", () => {
		pipeConn!.write(JSON.stringify({ type: "register_pipe" }) + "\n");
	});

	pipeConn.on("data", (chunk) => {
		pipeBuffer += chunk.toString();
		const lines = pipeBuffer.split("\n");
		pipeBuffer = lines.pop() || "";

		for (const line of lines) {
			if (!line.trim()) continue;
			try {
				const resp = JSON.parse(line);
				handlePipeResponse(pi, resp);
			} catch {
				// Skip malformed lines.
			}
		}
	});

	pipeConn.on("close", () => {
		pipeConn = null;
		for (const [guid, pending] of pendingCommands) {
			pending.reject(new Error("pipe disconnected"));
			pendingCommands.delete(guid);
		}
	});

	pipeConn.on("error", () => {
		pipeConn?.destroy();
		pipeConn = null;
	});
}

function handlePipeResponse(
	pi: ExtensionAPI,
	resp: { guid?: string; status?: string; success?: boolean; message?: string; data?: any; error?: string },
) {
	if (resp.status === "ok" && !resp.guid) return;

	if (!resp.guid) return;
	const pending = pendingCommands.get(resp.guid);
	if (!pending) return;

	if (resp.status === "update") {
		if (resp.message) {
			pi.sendMessage(
				{ customType: "pigo-update", content: resp.message, display: true },
				{ deliverAs: "steer" },
			);
		}
		return;
	}

	if (resp.status === "done" || resp.success) {
		pending.resolve(resp.data);
		pendingCommands.delete(resp.guid);
		return;
	}

	if (resp.status === "error" || resp.error) {
		pending.reject(new Error(resp.error || "command failed"));
		pendingCommands.delete(resp.guid);
		return;
	}
}

// -------------------------------------------------------------------
// Command execution (pipe with CLI fallback)
// -------------------------------------------------------------------

function generateGUID(): string {
	return "xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx".replace(/[xy]/g, (c) => {
		const r = (Math.random() * 16) | 0;
		return (c === "x" ? r : (r & 0x3) | 0x8).toString(16);
	});
}

async function runCommand(pi: ExtensionAPI, command: string, args: Record<string, any>): Promise<any> {
	if (pipeConn && !pipeConn.destroyed) {
		const guid = generateGUID();
		return new Promise((resolve, reject) => {
			const timeout = setTimeout(() => {
				pendingCommands.delete(guid);
				reject(new Error("command timed out"));
			}, 30000);

			pendingCommands.set(guid, {
				resolve: (v) => {
					clearTimeout(timeout);
					resolve(v);
				},
				reject: (e) => {
					clearTimeout(timeout);
					reject(e);
				},
			});

			pipeConn!.write(JSON.stringify({ guid, command, args }) + "\n");
		});
	}

	return runViaCLI(pi, command, args);
}

async function runViaCLI(pi: ExtensionAPI, command: string, args: Record<string, any>): Promise<any> {
	// For CLI fallback, use --json to get structured output.
	const cliArgs: string[] = [];

	if (command === "vault.search") {
		cliArgs.push("vault", "search", args.q || "", "--json");
	} else if (command === "vault.read") {
		cliArgs.push("vault", "read", args.id || "", "--json");
	} else if (command === "vault.write") {
		cliArgs.push("vault", "write", "--title", args.title || "", "--json");
		if (args.tags) cliArgs.push("--tags", (args.tags as string[]).join(","));
		if (args.body) cliArgs.push("--body", args.body);
	} else if (command === "vault.edit") {
		cliArgs.push("vault", "edit", args.id || "", "--json");
		if (args.body) cliArgs.push("--body", args.body);
		if (args.tags) cliArgs.push("--tags", (args.tags as string[]).join(","));
	} else if (command === "vault.list") {
		cliArgs.push("vault", "list", "--json");
	} else if (command === "vault.links") {
		cliArgs.push("vault", "links", args.id || "", "--json");
	} else if (command === "vault.tags") {
		cliArgs.push("vault", "tags", "--json");
	} else if (command === "vault.import") {
		cliArgs.push("vault", "import", args.url || "", "--json");
		if (args.tags) cliArgs.push("--tags", (args.tags as string[]).join(","));
	} else if (command === "web.search") {
		cliArgs.push("web", "search", args.q || "", "--json");
	} else {
		throw new Error(`unsupported CLI fallback for: ${command}`);
	}

	const result = await pi.exec("pigo", cliArgs);
	if (result.code !== 0) {
		throw new Error(result.stderr || `pigo exited with code ${result.code}`);
	}

	try {
		return JSON.parse(result.stdout);
	} catch {
		return result.stdout;
	}
}

// -------------------------------------------------------------------
// Tool registration
// -------------------------------------------------------------------

// truncateOutput wraps tool result text with pi's standard truncation so a
// very large vault result can't overflow the LLM's context. Uses truncateHead
// because for all our read-ish tools the most relevant content is at the top
// (search results ranked by score, list sorted by recency, etc.).
function truncateOutput(text: string, hint = "Use a more specific query or read individual notes."): string {
	const r = truncateHead(text, { maxLines: DEFAULT_MAX_LINES, maxBytes: DEFAULT_MAX_BYTES });
	if (!r.truncated) return r.content;
	return `${r.content}\n\n[Output truncated: ${r.outputLines}/${r.totalLines} lines, ${formatSize(r.outputBytes)}/${formatSize(r.totalBytes)}. ${hint}]`;
}

function registerVaultSearch(pi: ExtensionAPI) {
	pi.registerTool({
		name: "vault_search",
		label: "Vault Search",
		description: "Search the pigo knowledge vault by meaning, keywords, or tags. Returns ranked results (semantic + fuzzy merged).",
		promptSnippet: "vault_search(q, limit?) — search the persistent knowledge vault",
		promptGuidelines: [
			"Before answering a factual question about the user or this project, search the vault first — past conversations often answered it already.",
			"Start with a broad query, then refine. Vault search is cheap.",
		],
		parameters: Type.Object({
			q: Type.String({ description: "Search query" }),
			limit: Type.Optional(Type.Number({ description: "Max results (default 10)" })),
		}),
		async execute(_toolCallId, params, _signal, _onUpdate, _ctx) {
			const data = await runCommand(pi, "vault.search", params);
			return {
				content: [{ type: "text" as const, text: truncateOutput(formatSearchResults(data)) }],
				details: data,
			};
		},
	});
}

function registerVaultRead(pi: ExtensionAPI) {
	pi.registerTool({
		name: "vault_read",
		label: "Vault Read",
		description: "Read a note from the pigo vault by ID (slug).",
		promptSnippet: "vault_read(id) — read a full note from the knowledge vault",
		parameters: Type.Object({
			id: Type.String({ description: "Note ID (slug)" }),
		}),
		async execute(_toolCallId, params, _signal, _onUpdate, _ctx) {
			const data = await runCommand(pi, "vault.read", params);
			const text = typeof data === "string" ? data : data?.RawContent || JSON.stringify(data);
			return {
				content: [{ type: "text" as const, text: truncateOutput(text, "Ask the user for a more specific section or rely on vault_search to narrow down.") }],
				details: data,
			};
		},
	});
}

function registerVaultWrite(pi: ExtensionAPI) {
	pi.registerTool({
		name: "vault_write",
		label: "Vault Write",
		description: "Create a new note in the pigo vault. The note is indexed, embedded, and committed to git automatically.",
		promptSnippet: "vault_write(title, body, tags?) — save durable knowledge to the vault",
		promptGuidelines: [
			"When you learn something worth remembering across sessions — a decision, a pattern, a preference — save it with vault_write.",
			"Call vault_tags first and reuse existing tags; avoid creating near-duplicates like 'api' and 'apis'.",
			"Link related notes with [[slug]] syntax in the body so the knowledge graph grows.",
			"Check vault_search first to avoid duplicating a note that already exists.",
		],
		parameters: Type.Object({
			title: Type.String({ description: "Note title" }),
			body: Type.String({ description: "Note body (markdown)" }),
			tags: Type.Optional(Type.Array(Type.String(), { description: "Tags for categorization" })),
		}),
		// Precedent hook for future schema migrations. If the vault.write arg
		// shape ever changes, we can translate old arguments from resumed
		// sessions here without breaking the public parameter schema.
		prepareArguments(args) {
			return args;
		},
		async execute(_toolCallId, params, _signal, _onUpdate, _ctx) {
			const data = await runCommand(pi, "vault.write", params);
			const text = typeof data === "string" ? data : `Created note: ${data?.ID || "unknown"}`;
			return {
				content: [{ type: "text" as const, text }],
				details: data,
			};
		},
	});
}

function registerVaultEdit(pi: ExtensionAPI) {
	pi.registerTool({
		name: "vault_edit",
		label: "Vault Edit",
		description: "Update an existing note in the pigo vault. Provide new body and/or tags. Re-indexes, re-embeds, commits.",
		promptSnippet: "vault_edit(id, body?, tags?) — update an existing vault note",
		promptGuidelines: [
			"Prefer vault_edit over vault_write when an existing note covers the topic — preserves links, tags, and history.",
			"When in doubt, vault_read the current content first so your edit is additive rather than overwriting.",
		],
		parameters: Type.Object({
			id: Type.String({ description: "Note ID (slug)" }),
			body: Type.Optional(Type.String({ description: "New body content" })),
			tags: Type.Optional(Type.Array(Type.String(), { description: "New tags" })),
		}),
		async execute(_toolCallId, params, _signal, _onUpdate, _ctx) {
			const data = await runCommand(pi, "vault.edit", params);
			const text = typeof data === "string" ? data : `Updated note: ${params.id}`;
			return {
				content: [{ type: "text" as const, text }],
				details: data,
			};
		},
	});
}

function registerVaultList(pi: ExtensionAPI) {
	pi.registerTool({
		name: "vault_list",
		label: "Vault List",
		description: "List all notes in the pigo vault with titles, tags, and dates. Sorted by recency.",
		promptSnippet: "vault_list() — browse everything in the vault",
		promptGuidelines: [
			"Prefer vault_search over vault_list once the vault has more than a few dozen notes — search is more focused.",
		],
		parameters: Type.Object({}),
		async execute(_toolCallId, _params, _signal, _onUpdate, _ctx) {
			const data = await runCommand(pi, "vault.list", {});
			return {
				content: [{ type: "text" as const, text: truncateOutput(formatListResults(data)) }],
				details: data,
			};
		},
	});
}

function registerVaultLinks(pi: ExtensionAPI) {
	pi.registerTool({
		name: "vault_links",
		label: "Vault Links",
		description: "Show all connections for a note: relates_to (auto-discovered), links_to ([[wiki-links]]), and backlinks.",
		promptSnippet: "vault_links(id) — discover related notes through the knowledge graph",
		promptGuidelines: [
			"After reading an interesting note, call vault_links to surface related notes worth reading too.",
			"Backlinks often point at more context — follow them when the user's question is under-specified.",
		],
		parameters: Type.Object({
			id: Type.String({ description: "Note ID (slug)" }),
		}),
		async execute(_toolCallId, params, _signal, _onUpdate, _ctx) {
			const data = await runCommand(pi, "vault.links", params);
			const text = typeof data === "string" ? data : formatLinksResults(data);
			return {
				content: [{ type: "text" as const, text }],
				details: data,
			};
		},
	});
}

function registerVaultTags(pi: ExtensionAPI) {
	pi.registerTool({
		name: "vault_tags",
		label: "Vault Tags",
		description: "List all tags used in the vault with note counts. Use before tagging to stay consistent.",
		promptSnippet: "vault_tags() — list existing tags before creating new ones",
		promptGuidelines: [
			"Always call vault_tags before creating a note with tags — reuse existing ones rather than inventing near-duplicates.",
		],
		parameters: Type.Object({}),
		async execute(_toolCallId, _params, _signal, _onUpdate, _ctx) {
			const data = await runCommand(pi, "vault.tags", {});
			const text = typeof data === "string" ? data : formatTagResults(data);
			return {
				content: [{ type: "text" as const, text: truncateOutput(text, "The most-used tags are at the top; that's usually enough.") }],
				details: data,
			};
		},
	});
}

function registerWebSearch(pi: ExtensionAPI) {
	pi.registerTool({
		name: "web_search",
		label: "Web Search",
		description: "Search the web via SearXNG. Returns titles, URLs, and snippets.",
		promptSnippet: "web_search(q, limit?) — search the web, then vault_import useful results",
		promptGuidelines: [
			"When a web result is worth keeping for later, follow up with pigo_command(command='vault.import', args={url: '...'}) so it lives in the vault with git history.",
		],
		parameters: Type.Object({
			q: Type.String({ description: "Search query" }),
			limit: Type.Optional(Type.Number({ description: "Max results (default 10)" })),
		}),
		async execute(_toolCallId, params, _signal, _onUpdate, _ctx) {
			const data = await runCommand(pi, "web.search", params);
			const text = typeof data === "string" ? data : formatWebResults(data);
			return {
				content: [{ type: "text" as const, text: truncateOutput(text, "Refine the query for fewer, more specific results.") }],
				details: data,
			};
		},
	});
}

function registerPigoCommand(pi: ExtensionAPI) {
	pi.registerTool({
		name: "pigo_command",
		label: "Pigo Command",
		description: "Execute any registered pigo command by name. The escape hatch for commands without a dedicated tool.",
		promptSnippet: "pigo_command(command, args?) — call any pigo command by name",
		promptGuidelines: [
			"Start with command='system.methods' if you don't know what commands exist.",
			"Use this for vault.import (saving a URL to the vault) and for any new pigo commands that don't yet have dedicated tools.",
		],
		parameters: Type.Object({
			command: Type.String({ description: "Command name (e.g. system.methods, vault.import)" }),
			args: Type.Optional(Type.Record(Type.String(), Type.Any(), { description: "Command arguments" })),
		}),
		async execute(_toolCallId, params, _signal, _onUpdate, _ctx) {
			const data = await runCommand(pi, params.command, params.args || {});
			const text = typeof data === "string" ? data : JSON.stringify(data, null, 2);
			return {
				content: [{ type: "text" as const, text: truncateOutput(text) }],
				details: data,
			};
		},
	});
}

// -------------------------------------------------------------------
// Formatters
// -------------------------------------------------------------------

function formatSearchResults(data: any): string {
	if (typeof data === "string") return data;
	const results = data?.Results || data;
	if (!Array.isArray(results) || results.length === 0) return "No results found.";

	return results
		.map((r: any) => {
			const heading = r.Heading ? ` > ${r.Heading}` : "";
			const score = typeof r.Score === "number" ? `[${r.Score.toFixed(2)}] ` : "";
			return `${score}${r.Title}${heading}`;
		})
		.join("\n");
}

function formatSearchSummary(data: any): string {
	const results = data?.Results || data;
	if (!Array.isArray(results)) return "no results";
	return `${results.length} result${results.length === 1 ? "" : "s"}`;
}

function formatListResults(data: any): string {
	if (typeof data === "string") return data;
	if (!Array.isArray(data) || data.length === 0) return "Vault is empty.";

	return data
		.map((item: any) => {
			const tags = item.Tags?.length ? `  [${item.Tags.join(", ")}]` : "";
			return `${item.UpdatedAt}  ${item.Title}${tags}`;
		})
		.join("\n");
}

function formatLinksResults(data: any): string {
	if (typeof data === "string") return data;
	const lines: string[] = [`Links for: ${data.note_id}`, ""];
	lines.push(`  Relates to:  ${data.relates_to?.length ? data.relates_to.join(", ") : "(none)"}`);
	lines.push(`  Links to:    ${data.links_to?.length ? data.links_to.join(", ") : "(none)"}`);
	lines.push(`  Backlinks:   ${data.backlinks?.length ? data.backlinks.join(", ") : "(none)"}`);
	return lines.join("\n");
}

function formatTagResults(data: any): string {
	if (typeof data === "string") return data;
	if (!Array.isArray(data) || data.length === 0) return "No tags in vault.";

	return data.map((t: any) => `  ${String(t.count).padStart(3)}  ${t.tag}`).join("\n");
}

function formatWebResults(data: any): string {
	if (typeof data === "string") return data;
	if (!Array.isArray(data) || data.length === 0) return "No results found.";

	return data
		.map((r: any, i: number) => {
			let entry = `${i + 1}. ${r.title}\n   ${r.url}`;
			if (r.content) {
				const snippet = r.content.replace(/\n/g, " ").slice(0, 120);
				entry += `\n   ${snippet}${r.content.length > 120 ? "..." : ""}`;
			}
			return entry;
		})
		.join("\n\n");
}
