// task_review.go — CLI "task review": open a local browser UI for reviewing a task execution pack.
package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"html"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/davidl71/exarp-go/internal/framework"
)

func handleTaskReview(server framework.MCPServer, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("task review requires a task ID")
	}

	taskID := strings.TrimSpace(args[0])
	if taskID == "" {
		return fmt.Errorf("task review requires a task ID")
	}

	packJSON, err := loadExecutionPackJSON(server, taskID)
	if err != nil {
		return err
	}

	reviewHTML, err := buildTaskReviewHTML(taskID, packJSON)
	if err != nil {
		return err
	}

	path, err := writeTempHTML("exarp-task-review-", reviewHTML)
	if err != nil {
		return err
	}

	if err := openInBrowser(path); err != nil && !CLIOutputOpts.Quiet {
		_, _ = fmt.Fprintf(os.Stderr, "Warning: failed to open browser: %v\n", err)
	}

	if !CLIOutputOpts.Quiet {
		_, _ = fmt.Fprintf(os.Stdout, "Review UI written to: %s\n", path)
	}

	return nil
}

func loadExecutionPackJSON(server framework.MCPServer, taskID string) ([]byte, error) {
	ctx := context.Background()

	uri := fmt.Sprintf("stdio://agent/task/%s/execution-pack", taskID)
	toolArgs := map[string]interface{}{"uri": uri}
	argsBytes, err := json.Marshal(toolArgs)
	if err != nil {
		return nil, fmt.Errorf("task review: marshal read_resource args: %w", err)
	}

	result, err := server.CallTool(ctx, "read_resource", argsBytes)
	if err != nil {
		return nil, fmt.Errorf("task review: read_resource(%s): %w", uri, err)
	}
	if len(result) == 0 || strings.TrimSpace(result[0].Text) == "" {
		return nil, fmt.Errorf("task review: empty execution pack for %s", taskID)
	}

	// Ensure it’s valid JSON and also pretty-print for readability in UI.
	var parsed any
	if err := json.Unmarshal([]byte(result[0].Text), &parsed); err != nil {
		return nil, fmt.Errorf("task review: execution pack is not valid JSON: %w", err)
	}
	pretty, _ := json.MarshalIndent(parsed, "", "  ")
	return pretty, nil
}

func buildTaskReviewHTML(taskID string, packJSON []byte) (string, error) {
	exePath, _ := os.Executable()
	exePath = strings.TrimSpace(exePath)
	if exePath == "" {
		exePath = "exarp-go"
	}

	escapedJSON := html.EscapeString(string(packJSON))
	escapedTaskID := html.EscapeString(taskID)
	escapedExe := html.EscapeString(exePath)

	// This UI is intentionally local-file-only (no server). It generates the exact command
	// a reviewer can run to apply an approval result via task_workflow.
	page := fmt.Sprintf(`<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8"/>
  <meta name="viewport" content="width=device-width, initial-scale=1"/>
  <title>exarp-go task review: %s</title>
  <style>
    :root{
      --bg:#0b1020;
      --panel:#111831;
      --panel2:#0f1630;
      --text:#e8eefc;
      --muted:#a9b4d0;
      --accent:#7aa2ff;
      --good:#32d296;
      --bad:#ff6b6b;
      --border:rgba(122,162,255,.25);
      --mono: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", "Courier New", monospace;
      --sans: ui-sans-serif, system-ui, -apple-system, Segoe UI, Roboto, Helvetica, Arial, "Apple Color Emoji", "Segoe UI Emoji";
    }
    body{ margin:0; background:radial-gradient(1200px 700px at 20%% 10%%, rgba(122,162,255,.18), transparent 60%%), var(--bg); color:var(--text); font-family:var(--sans); }
    .wrap{ max-width:1100px; margin:32px auto; padding:0 18px 60px; }
    .hdr{ display:flex; gap:14px; align-items:flex-end; justify-content:space-between; flex-wrap:wrap; }
    h1{ font-size:22px; margin:0; letter-spacing:.2px; }
    .sub{ color:var(--muted); font-size:13px; margin-top:6px; }
    .grid{ display:grid; grid-template-columns: 1.2fr .8fr; gap:14px; margin-top:18px; }
    @media (max-width: 980px){ .grid{ grid-template-columns:1fr; } }
    .card{ background:linear-gradient(180deg, rgba(255,255,255,.03), transparent 60%%), var(--panel); border:1px solid var(--border); border-radius:14px; padding:14px; }
    .card h2{ margin:0 0 10px; font-size:14px; color:var(--muted); font-weight:600; letter-spacing:.12em; text-transform:uppercase; }
    textarea{ width:100%%; min-height:120px; resize:vertical; background:var(--panel2); color:var(--text); border:1px solid rgba(255,255,255,.08); border-radius:12px; padding:12px; font-family:var(--sans); font-size:14px; line-height:1.35; }
    pre{ margin:0; background:var(--panel2); border:1px solid rgba(255,255,255,.08); border-radius:12px; padding:12px; overflow:auto; font-family:var(--mono); font-size:12.5px; line-height:1.35; }
    .row{ display:flex; gap:10px; flex-wrap:wrap; align-items:center; }
    .btn{ border:1px solid rgba(255,255,255,.14); background:rgba(255,255,255,.04); color:var(--text); padding:10px 12px; border-radius:12px; cursor:pointer; font-weight:600; letter-spacing:.2px; }
    .btn.good{ border-color:rgba(50,210,150,.4); background:rgba(50,210,150,.10); }
    .btn.bad{ border-color:rgba(255,107,107,.4); background:rgba(255,107,107,.10); }
    .btn:hover{ filter:brightness(1.06); }
    .pill{ font-family:var(--mono); font-size:12px; color:var(--muted); }
    .hint{ color:var(--muted); font-size:13px; line-height:1.35; }
    .small{ font-size:12px; color:var(--muted); }
  </style>
</head>
<body>
  <div class="wrap">
    <div class="hdr">
      <div>
        <h1>Task review</h1>
        <div class="sub">
          Task: <span class="pill">%s</span>
        </div>
      </div>
      <div class="small">
        Generated by: <span class="pill">%s</span>
      </div>
    </div>

    <div class="grid">
      <div class="card">
        <h2>Execution pack (read-only)</h2>
        <pre id="pack">%s</pre>
      </div>

      <div class="card">
        <h2>Decision</h2>
        <div class="hint">
          Enter feedback (optional). Then click Approve/Reject to generate the exact command to apply the decision via
          <span class="pill">task_workflow action=apply_approval_result</span>.
        </div>
        <div style="height:10px"></div>
        <textarea id="feedback" placeholder="Optional feedback (e.g., what to change, what to verify)."></textarea>
        <div style="height:10px"></div>
        <div class="row">
          <button class="btn good" onclick="makeCmd('approved')">Approve</button>
          <button class="btn bad" onclick="makeCmd('rejected')">Reject</button>
          <button class="btn" onclick="copyCmd()">Copy</button>
        </div>
        <div style="height:10px"></div>
        <pre id="cmd"></pre>
        <div style="height:10px"></div>
        <div class="hint">
          Note: this UI does not execute commands. Run the generated command in a terminal.
        </div>
      </div>
    </div>
  </div>

<script>
  const taskId = %q;
  const exe = %q;
  let lastCmd = "";

  function makeCmd(result) {
    const feedback = document.getElementById("feedback").value || "";
    const args = { action: "apply_approval_result", task_id: taskId, result };
    if (feedback.trim() !== "") args.feedback = feedback;
    const json = JSON.stringify(args);
    lastCmd = exe + " -tool task_workflow -args '" + json + "'";
    document.getElementById("cmd").textContent = lastCmd;
  }

  async function copyCmd() {
    if (!lastCmd) {
      makeCmd("approved");
    }
    try {
      await navigator.clipboard.writeText(lastCmd);
    } catch (e) {
      alert("Copy failed; select and copy manually.");
    }
  }

  // Default render a command for convenience.
  makeCmd("approved");
</script>
</body>
</html>`,
		escapedTaskID, escapedTaskID, escapedExe, escapedJSON,
		taskID, exePath,
	)

	return page, nil
}

func writeTempHTML(prefix, contents string) (string, error) {
	f, err := os.CreateTemp("", prefix+"*.html")
	if err != nil {
		return "", fmt.Errorf("task review: create temp file: %w", err)
	}
	defer func() { _ = f.Close() }()

	if _, err := f.WriteString(contents); err != nil {
		return "", fmt.Errorf("task review: write temp file: %w", err)
	}
	return f.Name(), nil
}

func openInBrowser(path string) error {
	abs, err := filepath.Abs(path)
	if err == nil {
		path = abs
	}

	switch runtime.GOOS {
	case "darwin":
		return exec.Command("open", path).Start()
	case "windows":
		// "start" is a built-in shell command; must run via cmd.exe
		return exec.Command("cmd", "/c", "start", "", path).Start()
	default:
		return exec.Command("xdg-open", path).Start()
	}
}
