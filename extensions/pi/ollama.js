/**
 * Ollama Provider Extension for Pi
 *
 * Registers local Ollama models as a provider in the pi coding harness,
 * so you can use local models alongside cloud providers.
 *
 * Also provides slash commands and tools for model management:
 *   /ollama-list, /ollama-ps, /ollama-show, /ollama-pull, /ollama-delete, /ollama-config
 *
 * Ships with pigo. Install via: pigo install
 */

import { readFileSync, writeFileSync, mkdirSync, existsSync } from "fs";
import { join } from "path";
import { homedir } from "os";

export default async function (pi) {
  const configDir = join(homedir(), ".pi", "agent");
  const configFile = join(configDir, "models.json");

  // Default to localhost. Reads pigo config or pi models.json for overrides.
  let OLLAMA_HOST = "http://localhost:11434";

  // Try reading host from pigo config first, then pi models.json.
  const loadConfig = () => {
    // Check pigo config.
    try {
      const pigoConfig = join(homedir(), ".pigo", "config.toml");
      if (existsSync(pigoConfig)) {
        const text = readFileSync(pigoConfig, "utf-8");
        const match = text.match(/endpoint\s*=\s*"([^"]+)"/);
        if (match) {
          OLLAMA_HOST = match[1];
          return true;
        }
      }
    } catch {}

    // Fall back to pi models.json.
    try {
      if (existsSync(configFile)) {
        const config = JSON.parse(readFileSync(configFile, "utf-8"));
        if (config.providers?.ollama?.baseUrl) {
          OLLAMA_HOST = config.providers.ollama.baseUrl.replace("/v1", "");
          return true;
        }
      }
    } catch {}

    return false;
  };

  const saveConfig = (host, models = []) => {
    try {
      mkdirSync(configDir, { recursive: true });
      let config = {};
      if (existsSync(configFile))
        config = JSON.parse(readFileSync(configFile, "utf-8"));
      if (!config.providers) config.providers = {};
      if (!config.providers.ollama) config.providers.ollama = {};
      config.providers.ollama.baseUrl = `${host}/v1`;
      config.providers.ollama.api = "openai-completions";
      config.providers.ollama.apiKey = "ollama";
      config.providers.ollama.compat = {
        supportsDeveloperRole: false,
        supportsReasoningEffort: false,
      };
      if (models.length > 0) config.providers.ollama.models = models;
      writeFileSync(configFile, JSON.stringify(config, null, 2));
      return true;
    } catch {
      return false;
    }
  };

  // ── API helpers ──────────────────────────────────────────────────────────────

  const api = async (path, options = {}) => {
    const res = await fetch(`${OLLAMA_HOST}${path}`, options);
    if (!res.ok) throw new Error(`Ollama ${path} returned ${res.status}`);
    return res.json();
  };

  const streamPull = async (name, onChunk) => {
    const res = await fetch(`${OLLAMA_HOST}/api/pull`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ name }),
    });
    if (!res.ok) throw new Error(`Pull failed: ${res.status}`);

    const reader = res.body.getReader();
    const decoder = new TextDecoder();
    let buffer = "";

    while (true) {
      const { done, value } = await reader.read();
      if (done) break;
      buffer += decoder.decode(value, { stream: true });
      const lines = buffer.split("\n");
      buffer = lines.pop();
      for (const line of lines) {
        if (!line.trim()) continue;
        try {
          onChunk(JSON.parse(line));
        } catch {}
      }
    }
  };

  // ── Formatting ───────────────────────────────────────────────────────────────

  const fmtBytes = (bytes) => {
    if (!bytes) return "?";
    if (bytes >= 1e12) return `${(bytes / 1e12).toFixed(1)} TB`;
    if (bytes >= 1e9) return `${(bytes / 1e9).toFixed(1)} GB`;
    if (bytes >= 1e6) return `${(bytes / 1e6).toFixed(1)} MB`;
    return `${bytes} B`;
  };

  const fmtCaps = (caps) => {
    if (!caps?.length) return "";
    return `[${caps.join(", ")}]`;
  };

  // ── Model discovery ──────────────────────────────────────────────────────────

  const getModelDetails = async (modelName) => {
    try {
      const res = await fetch(`${OLLAMA_HOST}/api/show`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ name: modelName }),
      });
      if (!res.ok) return null;
      const data = await res.json();
      return data.error ? null : data;
    } catch {
      return null;
    }
  };

  const getOllamaModels = async () => {
    try {
      const data = await api("/api/tags");
      if (!data.models?.length) return [];

      const models = [];

      for (const model of data.models) {
        const details = await getModelDetails(model.name);

        // Skip models that don't support tool calling.
        if (details?.capabilities && !details.capabilities.includes("tools")) {
          continue;
        }

        let contextWindow = 4096;
        if (details?.model_info) {
          const key = Object.keys(details.model_info).find(
            (k) =>
              k.includes("context_length") ||
              k.includes("ctx_len") ||
              k.includes("max_seq_len"),
          );
          if (key) contextWindow = details.model_info[key];
        }
        if (contextWindow === 4096 && details?.parameters) {
          const m = details.parameters.match(/num_ctx\s+(\d+)/);
          if (m) contextWindow = parseInt(m[1], 10);
        }

        models.push({
          id: model.name,
          name: model.name.replace(/:[a-z0-9]+$/, "").replace(/-/g, " "),
          reasoning: false,
          input: ["text"],
          contextWindow,
          maxTokens: Math.floor(contextWindow * 0.25),
          cost: { input: 0, output: 0, cacheRead: 0, cacheWrite: 0 },
        });
      }

      return models;
    } catch {
      return [];
    }
  };

  const registerOllama = async () => {
    const models = await getOllamaModels();
    if (!models.length) return false;
    pi.registerProvider("ollama", {
      baseUrl: `${OLLAMA_HOST}/v1`,
      api: "openai-completions",
      apiKey: "ollama",
      compat: {
        supportsDeveloperRole: false,
        supportsReasoningEffort: false,
      },
      models,
    });
    return true;
  };

  // ── Slash commands ──────────────────────────────────────────────────────────

  pi.registerCommand("ollama-list", {
    description: "List all local Ollama models grouped by family",
    handler: async (_args, ctx) => {
      ctx.ui.notify("Fetching model list...", "info");
      let data;
      try {
        data = await api("/api/tags");
      } catch (e) {
        ctx.ui.notify(`Failed: ${e.message}`, "error");
        return;
      }

      const models = data.models ?? [];
      if (!models.length) {
        ctx.ui.notify("No models found", "info");
        return;
      }

      const groups = {};
      for (const m of models) {
        const details = await getModelDetails(m.name);
        const family = details?.details?.family ?? m.name.split(/[:/]/)[0];
        const caps = fmtCaps(details?.capabilities);
        const params = details?.details?.parameter_size ?? "";
        const quant = details?.details?.quantization_level ?? "";
        const size = fmtBytes(m.size);
        (groups[family] ??= []).push({ name: m.name, params, quant, size, caps });
      }

      const lines = [`Ollama Models (${models.length}) — ${OLLAMA_HOST}`, ""];
      for (const [family, entries] of Object.entries(groups)) {
        lines.push(family);
        for (const e of entries) {
          const meta = [e.params, e.quant, e.size].filter(Boolean).join(", ");
          lines.push(`  ${e.name}  ${meta}  ${e.caps}`);
        }
        lines.push("");
      }

      return lines.join("\n");
    },
  });

  pi.registerCommand("ollama-ps", {
    description: "Show models currently loaded in GPU memory",
    handler: async (_args, ctx) => {
      let data;
      try {
        data = await api("/api/ps");
      } catch (e) {
        ctx.ui.notify(`Failed: ${e.message}`, "error");
        return;
      }

      const models = data.models ?? [];
      if (!models.length) return "No models currently loaded in memory.";

      const lines = ["Running models:", ""];
      for (const m of models) {
        lines.push(`  ${m.name}`);
        lines.push(`    Size: ${fmtBytes(m.size)}  VRAM: ${fmtBytes(m.size_vram)}`);
        lines.push("");
      }
      return lines.join("\n");
    },
  });

  pi.registerCommand("ollama-show", {
    description: "Show detailed info for a model",
    handler: async (args, ctx) => {
      const name =
        (typeof args === "string" ? args : args?.[0] ?? "").trim() ||
        (await ctx.ui.input("Model name", ""));
      if (!name) return;

      let details;
      try {
        details = await getModelDetails(name);
      } catch (e) {
        ctx.ui.notify(`Failed: ${e.message}`, "error");
        return;
      }
      if (!details) {
        ctx.ui.notify(`Model not found: ${name}`, "error");
        return;
      }

      const info = details.model_info ?? {};
      const ctxKey = Object.keys(info).find((k) => k.includes("context_length"));
      const lines = [
        `Model: ${name}`,
        "",
        `  Family:       ${details.details?.family ?? "?"}`,
        `  Parameters:   ${details.details?.parameter_size ?? "?"}`,
        `  Quantization: ${details.details?.quantization_level ?? "?"}`,
        `  Format:       ${details.details?.format ?? "?"}`,
        `  Context:      ${ctxKey ? info[ctxKey].toLocaleString() : "?"} tokens`,
        `  Capabilities: ${fmtCaps(details.capabilities)}`,
      ];
      if (details.license) lines.push(`  License:      ${details.license.split("\n")[0]}`);
      return lines.join("\n");
    },
  });

  pi.registerCommand("ollama-pull", {
    description: "Download a model from Ollama registry",
    handler: async (args, ctx) => {
      const name =
        (typeof args === "string" ? args : args?.[0] ?? "").trim() ||
        (await ctx.ui.input("Model name to pull (e.g. llama3.2:8b)", ""));
      if (!name) return;

      ctx.ui.notify(`Pulling ${name}...`, "info");
      let lastPct = -1;

      try {
        await streamPull(name, (chunk) => {
          if (chunk.total && chunk.completed) {
            const pct = Math.floor((chunk.completed / chunk.total) * 100);
            if (pct !== lastPct && pct % 10 === 0) {
              lastPct = pct;
              ctx.ui.notify(`${name}: ${pct}% (${fmtBytes(chunk.completed)} / ${fmtBytes(chunk.total)})`, "info");
            }
          } else if (chunk.status && chunk.status !== "pulling manifest") {
            ctx.ui.notify(`${name}: ${chunk.status}`, "info");
          }
        });
        ctx.ui.notify(`${name} pulled successfully`, "success");
      } catch (e) {
        ctx.ui.notify(`Pull failed: ${e.message}`, "error");
      }
    },
  });

  pi.registerCommand("ollama-delete", {
    description: "Delete a local Ollama model",
    handler: async (args, ctx) => {
      const name =
        (typeof args === "string" ? args : args?.[0] ?? "").trim() ||
        (await ctx.ui.input("Model name to delete", ""));
      if (!name) return;

      const confirmed = await ctx.ui.confirm("Delete model?", `This will permanently delete ${name} from disk. Continue?`);
      if (!confirmed) return;

      try {
        const res = await fetch(`${OLLAMA_HOST}/api/delete`, {
          method: "DELETE",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ name }),
        });
        if (!res.ok) throw new Error(`${res.status}`);
        ctx.ui.notify(`Deleted ${name}`, "success");
      } catch (e) {
        ctx.ui.notify(`Delete failed: ${e.message}`, "error");
      }
    },
  });

  pi.registerCommand("ollama-config", {
    description: "Configure Ollama host and refresh models",
    handler: async (_args, ctx) => {
      const host = await ctx.ui.input(`Ollama host (current: ${OLLAMA_HOST})`, "");
      if (host === null || host === undefined) {
        ctx.ui.notify("Cancelled", "info");
        return;
      }
      if (host.trim()) OLLAMA_HOST = host.trim();

      ctx.ui.notify("Fetching models...", "info");
      const models = await getOllamaModels();
      if (!models.length) {
        ctx.ui.notify(`No models found at ${OLLAMA_HOST}`, "error");
        return;
      }

      if (saveConfig(OLLAMA_HOST, models)) {
        const ok = await registerOllama();
        ctx.ui.notify(
          ok ? `Ollama configured: ${models.length} models` : "Config saved but provider registration failed",
          ok ? "success" : "warning",
        );
      } else {
        ctx.ui.notify("Failed to save config", "error");
      }
    },
  });

  // ── Tools (AI-callable) ─────────────────────────────────────────────────────

  pi.registerTool({
    name: "ollama-list-models",
    label: "List Ollama Models",
    description: "List all locally available Ollama models with their sizes and capabilities",
    parameters: { type: "object", properties: {} },
    execute: async (_id, _params, _signal, _onUpdate) => {
      const data = await api("/api/tags");
      const models = (data.models ?? []).map((m) => ({
        name: m.name,
        size: fmtBytes(m.size),
        family: m.details?.family,
        params: m.details?.parameter_size,
        quantization: m.details?.quantization_level,
      }));
      const text = models.map((m) => `${m.name} — ${m.params ?? ""} ${m.family ?? ""} ${m.size}`).join("\n");
      return { content: [{ type: "text", text: text || "No models found" }], details: models };
    },
  });

  pi.registerTool({
    name: "ollama-running-models",
    label: "Running Ollama Models",
    description: "List Ollama models currently loaded in GPU memory",
    parameters: { type: "object", properties: {} },
    execute: async (_id, _params, _signal, _onUpdate) => {
      const data = await api("/api/ps");
      const models = data.models ?? [];
      const text = models.length
        ? models.map((m) => `${m.name} — ${fmtBytes(m.size_vram)} VRAM`).join("\n")
        : "No models currently loaded";
      return { content: [{ type: "text", text }], details: models };
    },
  });

  pi.registerTool({
    name: "ollama-pull-model",
    label: "Pull Ollama Model",
    description: "Download a model from the Ollama registry. Shows streaming progress.",
    parameters: {
      type: "object",
      properties: {
        name: { type: "string", description: "Model name to pull (e.g. llama3.2:8b)" },
      },
      required: ["name"],
    },
    execute: async (_id, params, _signal, onUpdate) => {
      const lines = [];
      let lastPct = -1;

      await streamPull(params.name, (chunk) => {
        let line;
        if (chunk.total && chunk.completed) {
          const pct = Math.floor((chunk.completed / chunk.total) * 100);
          if (pct !== lastPct && pct % 5 === 0) {
            lastPct = pct;
            line = `Downloading: ${pct}% (${fmtBytes(chunk.completed)} / ${fmtBytes(chunk.total)})`;
          }
        } else if (chunk.status) {
          line = chunk.status;
        }

        if (line) {
          lines.push(line);
          onUpdate?.({
            content: [{ type: "text", text: lines.join("\n") }],
            details: { status: chunk.status, chunk },
          });
        }
      });

      return {
        content: [{ type: "text", text: `Successfully pulled ${params.name}` }],
        details: { ok: true, model: params.name },
      };
    },
  });

  // ── Startup ─────────────────────────────────────────────────────────────────

  loadConfig();
  const ok = await registerOllama();
  if (ok) {
    console.log("[pigo/ollama] Registered Ollama provider");
  } else {
    console.log("[pigo/ollama] No tool-capable models found. Use /ollama-config to set host.");
  }
}
