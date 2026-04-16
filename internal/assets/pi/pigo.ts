/**
 * pigo — Pi Extension
 *
 * Gives the AI native tools for the pigo knowledge vault.
 * Connects to the pigo daemon via persistent TCP pipe for real-time streaming.
 * Falls back to CLI execution if the daemon isn't running.
 *
 * Install: copy or symlink this file into .pi/extensions/
 */

import type { ExtensionAPI } from "@mariozechner/pi-coding-agent";
import { Type } from "@sinclair/typebox";
import { createConnection, type Socket } from "net";
import { readFileSync, existsSync } from "fs";
import { homedir } from "os";
import { join } from "path";

// -------------------------------------------------------------------
// Config
// -------------------------------------------------------------------

const PIPE_HOST = "localhost";
const PIPE_PORT = 9877;
const SYSTEM_PROMPT_PATH = join(homedir(), ".pigo", "system.md");

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
	// Inject system prompt if it exists.
	pi.on("before_agent_start", async (event) => {
		try {
			if (existsSync(SYSTEM_PROMPT_PATH)) {
				const prompt = readFileSync(SYSTEM_PROMPT_PATH, "utf-8");
				return { systemPrompt: event.systemPrompt + "\n\n" + prompt };
			}
		} catch {
			// Silently skip if unreadable.
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
		// Keep the last (possibly incomplete) line in the buffer.
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
		// Reject any pending commands.
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
	// Pipe registration ack — not a command response.
	if (resp.status === "ok" && !resp.guid) return;

	if (!resp.guid) return;
	const pending = pendingCommands.get(resp.guid);
	if (!pending) return;

	if (resp.status === "update") {
		// Stream progress to the AI as a steer message.
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
	// Try pipe first.
	if (pipeConn && !pipeConn.destroyed) {
		const guid = generateGUID();
		return new Promise((resolve, reject) => {
			pendingCommands.set(guid, { resolve, reject });

			const timeout = setTimeout(() => {
				pendingCommands.delete(guid);
				reject(new Error("command timed out"));
			}, 30000);

			// Clear timeout when command completes.
			const origResolve = resolve;
			const origReject = reject;
			pendingCommands.set(guid, {
				resolve: (v) => {
					clearTimeout(timeout);
					origResolve(v);
				},
				reject: (e) => {
					clearTimeout(timeout);
					origReject(e);
				},
			});

			pipeConn!.write(JSON.stringify({ guid, command, args }) + "\n");
		});
	}

	// Fallback: CLI execution.
	return runViaCLI(pi, command, args);
}

async function runViaCLI(pi: ExtensionAPI, command: string, args: Record<string, any>): Promise<any> {
	const cliArgs = ["vault"]; // namespace

	// Map command to CLI subcommand + flags.
	if (command === "vault.search") {
		cliArgs.push("search", args.q || "");
	} else if (command === "vault.read") {
		cliArgs.push("read", args.id || "");
	} else if (command === "vault.write") {
		cliArgs.push("write", "--title", args.title || "");
		if (args.tags) cliArgs.push("--tags", (args.tags as string[]).join(","));
		if (args.body) cliArgs.push("--body", args.body);
	} else if (command === "vault.edit") {
		cliArgs.push("edit", args.id || "");
		if (args.body) cliArgs.push("--body", args.body);
		if (args.tags) cliArgs.push("--tags", (args.tags as string[]).join(","));
	} else if (command === "vault.list") {
		cliArgs.push("list");
	} else {
		throw new Error(`unsupported CLI fallback for: ${command}`);
	}

	const result = await pi.exec("pigo", cliArgs);
	if (result.code !== 0) {
		throw new Error(result.stderr || `pigo exited with code ${result.code}`);
	}
	return result.stdout;
}

// -------------------------------------------------------------------
// Tool registration
// -------------------------------------------------------------------

function registerVaultSearch(pi: ExtensionAPI) {
	pi.registerTool({
		name: "vault_search",
		label: "Vault Search",
		description: "Search the pigo knowledge vault by meaning, keywords, or tags. Returns ranked results.",
		promptSnippet: "vault_search(q) — search the persistent knowledge vault",
		parameters: Type.Object({
			q: Type.String({ description: "Search query" }),
			limit: Type.Optional(Type.Number({ description: "Max results (default 10)" })),
		}),
		async execute(_toolCallId, params, _signal, _onUpdate, _ctx) {
			const data = await runCommand(pi, "vault.search", params);
			return {
				content: [{ type: "text" as const, text: formatSearchResults(data) }],
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
		promptSnippet: "vault_read(id) — read a note from the knowledge vault",
		parameters: Type.Object({
			id: Type.String({ description: "Note ID (slug)" }),
		}),
		async execute(_toolCallId, params, _signal, _onUpdate, _ctx) {
			const data = await runCommand(pi, "vault.read", params);
			const text = typeof data === "string" ? data : (data?.RawContent || JSON.stringify(data));
			return {
				content: [{ type: "text" as const, text }],
				details: data,
			};
		},
	});
}

function registerVaultWrite(pi: ExtensionAPI) {
	pi.registerTool({
		name: "vault_write",
		label: "Vault Write",
		description: "Create a new note in the pigo vault. The note is indexed, embedded, and committed to git.",
		promptSnippet: "vault_write(title, body, tags?) — save knowledge to the vault",
		promptGuidelines: [
			"When you learn something worth remembering across sessions, save it with vault_write.",
			"Use descriptive titles and relevant tags so the note is findable later.",
		],
		parameters: Type.Object({
			title: Type.String({ description: "Note title" }),
			body: Type.String({ description: "Note body (markdown)" }),
			tags: Type.Optional(Type.Array(Type.String(), { description: "Tags for categorization" })),
		}),
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
		description: "Update an existing note in the pigo vault. Provide new body and/or tags.",
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
		description: "List all notes in the pigo vault with titles, tags, and dates.",
		parameters: Type.Object({}),
		async execute(_toolCallId, _params, _signal, _onUpdate, _ctx) {
			const data = await runCommand(pi, "vault.list", {});
			return {
				content: [{ type: "text" as const, text: formatListResults(data) }],
				details: data,
			};
		},
	});
}

function registerPigoCommand(pi: ExtensionAPI) {
	pi.registerTool({
		name: "pigo_command",
		label: "Pigo Command",
		description:
			"Execute any registered pigo command by name. Use system.methods to discover available commands.",
		parameters: Type.Object({
			command: Type.String({ description: "Command name (e.g. system.methods)" }),
			args: Type.Optional(Type.Record(Type.String(), Type.Any(), { description: "Command arguments" })),
		}),
		async execute(_toolCallId, params, _signal, _onUpdate, _ctx) {
			const data = await runCommand(pi, params.command, params.args || {});
			const text = typeof data === "string" ? data : JSON.stringify(data, null, 2);
			return {
				content: [{ type: "text" as const, text }],
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
	if (!Array.isArray(data) || data.length === 0) return "No results found.";

	return data
		.map((r: any) => {
			const heading = r.Heading ? ` > ${r.Heading}` : "";
			const score = typeof r.Score === "number" ? `[${r.Score.toFixed(2)}] ` : "";
			return `${score}${r.Title}${heading}`;
		})
		.join("\n");
}

function formatSearchSummary(data: any): string {
	if (!Array.isArray(data)) return "no results";
	return `${data.length} result${data.length === 1 ? "" : "s"}`;
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
