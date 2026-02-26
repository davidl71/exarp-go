/**
 * exarp-go OpenCode Plugin
 *
 * Complements the exarp-go MCP server with lifecycle hooks:
 * - Auto-injects PROJECT_ROOT into shell environment
 * - Injects task state into every system prompt (virtual sidebar)
 * - Injects task state into compaction so it survives context resets
 * - macOS notifications on session idle
 * - Slash commands: /tasks, /prime, /scorecard, /health
 */

import type { Plugin } from "@opencode-ai/plugin";

const EXARP_BINARY = process.env.EXARP_GO_BINARY || "exarp-go";
const CACHE_TTL_MS = 30_000;

interface TaskSummary {
  id: string;
  content: string;
  status: string;
  priority: string;
}

let cachedTaskContext: { text: string; expires: number } | null = null;

async function runExarp(
  $: any,
  tool: string,
  args: Record<string, unknown>
): Promise<string> {
  try {
    const argsJson = JSON.stringify(args);
    const result =
      await $`${EXARP_BINARY} -tool ${tool} -args ${argsJson}`.text();
    return result.trim();
  } catch {
    return "";
  }
}

async function getTaskSummary($: any): Promise<string> {
  const raw = await runExarp($, "task_workflow", {
    action: "sync",
    sub_action: "list",
    output_format: "json",
    compact: true,
  });

  if (!raw) return "No tasks available.";

  try {
    const lines = raw.split("\n");
    const jsonLine = lines.find((l: string) => l.trim().startsWith("{"));
    if (!jsonLine) return raw.slice(0, 500);

    const data = JSON.parse(jsonLine);
    const tasks: TaskSummary[] = data.tasks || [];
    if (tasks.length === 0) return "No tasks.";

    const byStatus: Record<string, TaskSummary[]> = {};
    for (const t of tasks) {
      const s = t.status || "Unknown";
      if (!byStatus[s]) byStatus[s] = [];
      byStatus[s].push(t);
    }

    const sections: string[] = [];
    for (const [status, items] of Object.entries(byStatus)) {
      sections.push(
        `**${status}** (${items.length}): ${items.map((t) => `${t.id} ${t.content}`).join("; ")}`
      );
    }
    return sections.join("\n");
  } catch {
    return raw.slice(0, 500);
  }
}

async function getCachedTaskContext($: any): Promise<string> {
  const now = Date.now();
  if (cachedTaskContext && now < cachedTaskContext.expires) {
    return cachedTaskContext.text;
  }
  const text = await getTaskSummary($);
  cachedTaskContext = { text, expires: now + CACHE_TTL_MS };
  return text;
}

export const ExarpGoPlugin: Plugin = async ({ $, directory }) => {
  const projectRoot = directory;

  return {
    "shell.env": async (_input, output) => {
      output.env.PROJECT_ROOT = projectRoot;
    },

    event: async ({ event }) => {
      if (event.type === "session.idle") {
        try {
          await $`osascript -e 'display notification "Session idle — check results" with title "exarp-go"'`.quiet();
        } catch {
          // non-macOS or osascript not available
        }
      }
      if (
        event.type === "todo.updated" ||
        event.type === "session.created"
      ) {
        cachedTaskContext = null;
      }
    },

    "experimental.chat.system.transform": async (_input, output) => {
      const taskContext = await getCachedTaskContext($);
      output.system.push(`
<exarp-go-context>
## Current Tasks

${taskContext}

You have access to exarp-go MCP tools: task_workflow, report, session, health, testing.
Use task_workflow to manage tasks. Use session(action="prime") for full project context.
Mark tasks Done when you complete work that satisfies them.
</exarp-go-context>
`);
    },

    "experimental.session.compacting": async (_input, output) => {
      const taskContext = await getCachedTaskContext($);
      output.context.push(`
## exarp-go Task State (injected by plugin)

${taskContext}

Use the exarp-go MCP tools (task_workflow, report, session) to manage tasks.
Always call session(action="prime") at the start of a new session.
`);
    },

    async config(config) {
      config.command = config.command ?? {};

      config.command["tasks"] = {
        description: "List current exarp-go tasks",
        template: `List my current tasks. Use the task_workflow MCP tool with action="sync", sub_action="list" to get all tasks grouped by status. Show a concise summary.`,
      };

      config.command["prime"] = {
        description: "Prime session with exarp-go context",
        template: `Prime the session. Call the session MCP tool with action="prime", include_hints=true, include_tasks=true. Then summarize the current project state, suggested next tasks, and any handoff notes.`,
      };

      config.command["scorecard"] = {
        description: "Show project scorecard",
        template: `Generate a project scorecard. Use the report MCP tool with action="scorecard". Show the results in a clear, readable format with scores and recommendations.`,
      };

      config.command["health"] = {
        description: "Run project health checks",
        template: `Run project health checks. Use the health MCP tool with action="tools" to check tool registration, then action="docs" for documentation health. Summarize findings.`,
      };
    },
  };
};
