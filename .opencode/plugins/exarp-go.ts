/**
 * exarp-go OpenCode Plugin
 *
 * Complements the exarp-go MCP server with lifecycle hooks:
 * - Auto-injects PROJECT_ROOT into shell environment
 * - Injects task state into system prompt + user messages (virtual sidebar)
 * - Skips injection for sub-agents and title generation (agent filtering)
 * - Injects task state into compaction so it survives context resets
 * - Custom tools: exarp_tasks, exarp_update_task, exarp_prime, exarp_config, exarp_followup
 * - tool.execute.after: appends in-progress task reminders to tool output
 * - todo.updated: refreshes task cache when tasks change
 * - tui.prompt.append: ready for follow-up injection
 * - macOS notifications on session idle; toasts on session errors
 * - Slash commands: /tasks, /prime, /scorecard, /health, /config
 * - Config tool for viewing and modifying exarp-go configuration
 * - Follow-up suggestions when tasks are completed
 */

import type { Plugin } from "@opencode-ai/plugin";
import { tool } from "@opencode-ai/plugin";

const EXARP_BINARY = process.env.EXARP_GO_BINARY || "exarp-go";
const CACHE_TTL_MS = 30_000;

interface TaskSummary {
  id: string;
  content: string;
  status: string;
  priority: string;
  tags?: string[];
}

interface TaskCache {
  text: string;
  tasks: TaskSummary[];
  expires: number;
}

let taskCache: TaskCache | null = null;
const seenSessions = new Set<string>();

async function runExarp(
  $: any,
  toolName: string,
  args: Record<string, unknown>
): Promise<string> {
  try {
    const argsJson = JSON.stringify(args);
    const result =
      await $`${EXARP_BINARY} -tool ${toolName} -args ${argsJson}`.text();
    return result.trim();
  } catch {
    return "";
  }
}

function parseTasks(raw: string): TaskSummary[] {
  if (!raw) return [];
  try {
    const lines = raw.split("\n");
    const jsonLine = lines.find((l: string) => l.trim().startsWith("{"));
    if (!jsonLine) return [];
    const data = JSON.parse(jsonLine);
    return data.tasks || [];
  } catch {
    return [];
  }
}

function formatTasks(tasks: TaskSummary[], status?: string): string {
  const filtered = status
    ? tasks.filter((t) => t.status.toLowerCase() === status.toLowerCase())
    : tasks;

  if (filtered.length === 0)
    return status ? `No ${status} tasks.` : "No tasks.";

  const byStatus: Record<string, TaskSummary[]> = {};
  for (const t of filtered) {
    const s = t.status || "Unknown";
    if (!byStatus[s]) byStatus[s] = [];
    byStatus[s].push(t);
  }

  const sections: string[] = [];
  for (const [s, items] of Object.entries(byStatus)) {
    const lines = items.map(
      (t) =>
        `  - ${t.id} [${t.priority}] ${t.content}${t.tags?.length ? ` (${t.tags.join(", ")})` : ""}`
    );
    sections.push(`**${s}** (${items.length}):\n${lines.join("\n")}`);
  }
  return sections.join("\n\n");
}

function formatTasksCompact(tasks: TaskSummary[]): string {
  if (tasks.length === 0) return "No tasks.";

  const byStatus: Record<string, TaskSummary[]> = {};
  for (const t of tasks) {
    const s = t.status || "Unknown";
    if (!byStatus[s]) byStatus[s] = [];
    byStatus[s].push(t);
  }

  const parts: string[] = [];
  for (const [s, items] of Object.entries(byStatus)) {
    parts.push(
      `${s}(${items.length}): ${items.map((t) => t.id).join(", ")}`
    );
  }
  return parts.join(" | ");
}

async function refreshCache($: any): Promise<TaskCache> {
  const now = Date.now();
  if (taskCache && now < taskCache.expires) return taskCache;

  const raw = await runExarp($, "task_workflow", {
    action: "sync",
    sub_action: "list",
    output_format: "json",
    compact: true,
  });

  const tasks = parseTasks(raw);
  const text = formatTasks(tasks);
  taskCache = { text, tasks, expires: now + CACHE_TTL_MS };
  return taskCache;
}

function invalidateCache() {
  taskCache = null;
}

type ToastVariant = "success" | "error" | "info" | "warning";

async function showToast(
  client: any,
  message: string,
  variant: ToastVariant = "info"
) {
  try {
    await client.tui.showToast({ body: { message, variant } });
  } catch {
    try {
      await client.app.log({
        body: { service: "exarp-go", level: "info", message },
      });
    } catch {
      // SDK not available (headless / non-TUI mode)
    }
  }
}

async function appendToPrompt(client: any, text: string) {
  try {
    await client.tui.appendPrompt({ body: { text } });
  } catch {
    // TUI not available
  }
}

function isSubAgent(input: any): boolean {
  const mode = input?.agent?.mode;
  return mode === "subagent" || mode === "title";
}

function getInProgressTasks(tasks: TaskSummary[]): TaskSummary[] {
  return tasks.filter((t) => t.status === "In Progress");
}

export const ExarpGoPlugin: Plugin = async ({ $, client, directory }) => {
  const projectRoot = directory;

  return {
    "shell.env": async (_input, output) => {
      output.env.PROJECT_ROOT = projectRoot;
    },

    tool: {
      exarp_tasks: tool({
        description:
          "List exarp-go project tasks. Optionally filter by status (Todo, In Progress, Review, Done). Faster than MCP round-trip.",
        args: {
          status: tool.schema
            .string()
            .optional()
            .describe(
              "Filter by status: Todo, In Progress, Review, Done. Omit for all."
            ),
        },
        async execute(args) {
          const cache = await refreshCache($);
          return formatTasks(cache.tasks, args.status);
        },
      }),

      exarp_update_task: tool({
        description:
          "Update an exarp-go task status. Use when completing work that satisfies a task.",
        args: {
          task_id: tool.schema
            .string()
            .describe("Task ID, e.g. T-1772113376070017000"),
          new_status: tool.schema
            .string()
            .describe("New status: Todo, In Progress, Review, or Done"),
        },
        async execute(args) {
          const result = await runExarp($, "task_workflow", {
            action: "update",
            task_id: args.task_id,
            new_status: args.new_status,
          });
          invalidateCache();
          return result || `Updated ${args.task_id} → ${args.new_status}`;
        },
      }),

      exarp_prime: tool({
        description:
          "Prime session with full exarp-go project context: tasks, hints, handoffs, suggested next actions.",
        args: {},
        async execute() {
          const result = await runExarp($, "session", {
            action: "prime",
            include_hints: true,
            include_tasks: true,
            compact: true,
          });
          return result || "Session primed (no data returned).";
        },
      }),

      exarp_config: tool({
        description:
          "Get or set exarp-go configuration values. Use action='get' to retrieve a value, action='set' to update, action='show' to display all config, action='reset' to reset to defaults, action='diff' to compare with defaults, action='history' to see change history, action='template' to list available templates.",
        args: {
          action: tool.schema
            .string()
            .describe("Action: get, set, show, reset, diff, history, template"),
          key: tool.schema
            .string()
            .optional()
            .describe("Config key (e.g., 'timeouts.task_lock_lease'). Required for get/set/reset."),
          value: tool.schema
            .string()
            .optional()
            .describe("Value to set (for set action)"),
        },
        async execute(args) {
          const { action, key, value } = args;
          if (action === "show" || action === undefined) {
            const result = await runExarp($, "config", {
              subcommand: "show",
              positional: ["yaml"],
            });
            return result || "Config (using defaults)";
          }
          if (action === "get") {
            if (!key) return "Error: 'key' required for get action";
            const result = await runExarp($, "config", {
              subcommand: "get",
              positional: [key],
            });
            return result.trim() || `No value for ${key}`;
          }
          if (action === "set") {
            if (!key || !value) return "Error: 'key' and 'value' required for set action";
            const result = await runExarp($, "config", {
              subcommand: "set",
              positional: [`${key}=${value}`],
            });
            return result.trim() || `Set ${key} = ${value}`;
          }
          if (action === "reset") {
            const target = key || "all";
            const result = await runExarp($, "config", {
              subcommand: "reset",
              positional: [target],
            });
            return result.trim() || `Reset ${target} to defaults`;
          }
          if (action === "diff") {
            const result = await runExarp($, "config", {
              subcommand: "diff",
              positional: [],
            });
            return result.trim() || "Config matches defaults";
          }
          if (action === "history") {
            const result = await runExarp($, "config", {
              subcommand: "history",
              positional: [],
            });
            return result.trim() || "No config history";
          }
          if (action === "template") {
            const result = await runExarp($, "config", {
              subcommand: "template",
              positional: key ? [key] : [],
            });
            return result.trim() || "Template action";
          }
          if (action === "reload") {
            const result = await runExarp($, "config", {
              subcommand: "reload",
              positional: [],
            });
            return result.trim() || "Config reloaded";
          }
          return "Error: action must be 'get', 'set', 'show', 'reset', 'diff', 'history', 'template', or 'reload'";
        },
      }),

      // Get follow-up suggestions for a completed task, or create follow-ups
      exarp_followup: tool({
        description:
          "Get AI-suggested follow-up tasks for a completed task, or create them. Use after completing a task to see what comes next.",
        args: {
          action: tool.schema
            .string()
            .describe("Action: suggest or create"),
          task_id: tool.schema
            .string()
            .optional()
            .describe("Task ID to get suggestions for (optional, uses most recent Done task if not provided)"),
          suggestions: tool.schema
            .string()
            .optional()
            .describe("JSON array of suggestions to create (for create action)"),
        },
        async execute(args) {
          const action = args.action || "suggest";
          
          if (action === "suggest") {
            // Call task_workflow to get suggestions (this would need MCP)
            // For now, use session to prime with context
            const result = await runExarp($, "session", {
              action: "prime",
              include_hints: true,
              include_tasks: true,
            });
            return result || "No follow-up suggestions available. Make sure a task was recently completed.";
          }
          
          if (action === "create") {
            if (!args.suggestions) return "Error: 'suggestions' JSON array required for create action";
            // This would call task_workflow create for each suggestion
            return "Creating follow-ups... (use task_workflow MCP tool with action=create to create individual tasks)";
          }
          
          return "Error: action must be 'suggest' or 'create'";
        },
      }),
    },

    event: async ({ event }) => {
      if (event.type === "session.idle") {
        try {
          await $`osascript -e 'display notification "Session idle — check results" with title "exarp-go"'`.quiet();
        } catch {
          // non-macOS or osascript not available
        }
      }

      if (event.type === "session.created") {
        invalidateCache();
        const cache = await refreshCache($);
        const todo = cache.tasks.filter((t) => t.status === "Todo").length;
        const inProgress = cache.tasks.filter(
          (t) => t.status === "In Progress"
        ).length;
        if (cache.tasks.length > 0) {
          await showToast(
            client,
            `exarp-go: ${cache.tasks.length} tasks (${todo} todo, ${inProgress} in progress)`,
            "success"
          );
        }
      }

      if (event.type === "todo.updated") {
        invalidateCache();
        await showToast(
          client,
          "exarp-go: tasks updated",
          "info"
        );
      }

      if (event.type === "tui.command.execute") {
        const props = (event as any).properties || {};
        const cmd = props.command || "";
        if (
          cmd === "tasks" ||
          cmd === "prime" ||
          cmd === "scorecard" ||
          cmd === "health"
        ) {
          invalidateCache();
        }
      }

      if (event.type === "tui.prompt.append") {
        const props = (event as any).properties || {};
        const text: string = props.text || "";
        const taskIds = text.match(/T-\d{10,}/g);
        if (taskIds && taskIds.length > 0) {
          const cache = await refreshCache($);
          const matched = taskIds
            .map((id) => cache.tasks.find((t) => t.id === id))
            .filter(Boolean) as TaskSummary[];
          if (matched.length > 0) {
            const context = matched
              .map(
                (t) => `${t.id} [${t.status}/${t.priority}]: ${t.content}`
              )
              .join("\n");
            await appendToPrompt(
              client,
              `\n\n[Task context:\n${context}]`
            );
          }
        }
      }

      if (event.type === "tui.toast.show") {
        const props = (event as any).properties || {};
        const msg: string = props.message || props.text || "";
        if (msg.toLowerCase().includes("task") && !msg.includes("exarp")) {
          invalidateCache();
        }
      }

      if (event.type === "session.error") {
        const props = (event as any).properties || {};
        const errMsg: string = props.message || props.error || "Unknown error";
        await showToast(client, `exarp-go: session error — ${errMsg}`, "error");
      }
    },

    "chat.message": async (input, output) => {
      if (isSubAgent(input)) return;
      const isFirstMessage = !seenSessions.has(input.sessionID);
      if (isFirstMessage) {
        seenSessions.add(input.sessionID);
        const cache = await refreshCache($);
        if (cache.tasks.length > 0) {
          output.parts.unshift({
            type: "text",
            text: `[exarp-go tasks: ${formatTasksCompact(cache.tasks)}]`,
            synthetic: true,
          } as any);
        }
      }
    },

    "experimental.chat.system.transform": async (input, output) => {
      if (isSubAgent(input)) return;
      const cache = await refreshCache($);
      output.system.push(`
<exarp-go-context>
## Current Tasks

${cache.text}

You have exarp-go tools available:
- exarp_tasks: Quick task list (plugin tool, no MCP needed)
- exarp_update_task: Update task status (plugin tool)
- exarp_prime: Full session prime with tasks, hints, handoffs
- exarp_config: Get/set config values (action=get|set|show, key=..., value=...)
- exarp_followup: Get/create follow-up tasks (action=suggest|create, task_id=...)
- task_workflow, report, session, health: MCP tools for advanced operations
When completing a task (set status to Done), check for follow-up suggestions and create them to maintain momentum.
</exarp-go-context>
`);
    },

    "tool.execute.after": async (_input, output) => {
      const cache = await refreshCache($);
      const active = getInProgressTasks(cache.tasks);
      if (active.length > 0) {
        const reminder = active
          .map((t) => `${t.id}: ${t.content}`)
          .join("; ");
        output.result += `\n\n[In Progress (${active.length}): ${reminder}]`;
      }
    },

    "experimental.session.compacting": async (_input, output) => {
      const cache = await refreshCache($);
      output.context.push(`
## exarp-go Task State (injected by plugin)

${cache.text}

Use exarp_tasks/exarp_update_task (plugin tools) or task_workflow (MCP) to manage tasks.
Always call exarp_prime or session(action="prime") at the start of a new session.
`);
    },

    // Show toast when tasks are updated
    "todo.updated": async (input: any) => {
      // Refresh cache when tasks change
      taskCache = null;
      await refreshCache($);
    },

    // Append follow-up suggestions to prompt when relevant
    "tui.prompt.append": async (input: any, output: any) => {
      // Could inject follow-up suggestions here if we detect a task was just completed
      // For now, we keep the system context updated via other hooks
    },

    async config(config) {
      config.experimental = config.experimental ?? {};
      config.experimental.primary_tools = [
        ...(config.experimental.primary_tools || []),
        "exarp_update_task",
      ];

      config.command = config.command ?? {};

      config.command["tasks"] = {
        description: "List current exarp-go tasks",
        template: `List my current tasks. Use the exarp_tasks tool to get all tasks grouped by status. Show a concise summary.`,
      };

      config.command["prime"] = {
        description: "Prime session with exarp-go context",
        template: `Prime the session. Call the exarp_prime tool. Then summarize the current project state, suggested next tasks, and any handoff notes.`,
      };

      config.command["scorecard"] = {
        description: "Show project scorecard",
        template: `Generate a project scorecard. Use the report MCP tool with action="scorecard". Show the results in a clear, readable format with scores and recommendations.`,
      };

      config.command["health"] = {
        description: "Run project health checks",
        template: `Run project health checks. Use the health MCP tool with action="tools" to check tool registration, then action="docs" for documentation health. Summarize findings.`,
      };

      config.command["config"] = {
        description: "Show exarp-go configuration",
        template: `Show the current exarp-go configuration. Use the exarp_config tool with action="show" to display all config values in YAML format.`,
      };

      config.command["followup"] = {
        description: "Get follow-up task suggestions",
        template: `Get AI-suggested follow-up tasks. Use exarp_followup tool with action="suggest" to get suggestions for the most recently completed task. Review suggestions and use task_workflow MCP to create them.`,
      };
    },
  };
};
