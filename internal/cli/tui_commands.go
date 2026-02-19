// tui_commands.go — TUI tea.Cmd factories for background operations.
package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/davidl71/exarp-go/internal/config"
	"github.com/davidl71/exarp-go/internal/framework"
	"github.com/davidl71/exarp-go/internal/queue"
	"github.com/davidl71/exarp-go/internal/tools"
)

func tick() tea.Cmd {
	refreshInterval := 5 * time.Second

	return tea.Tick(refreshInterval, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func loadTasks(server framework.MCPServer, status string) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		tasks, err := listTasksViaMCP(ctx, server, status)
		return taskLoadedMsg{tasks: tasks, err: err}
	}
}

// updateTaskStatusCmd updates a task's status via task_workflow MCP tool and returns statusUpdateDoneMsg.
func updateTaskStatusCmd(server framework.MCPServer, taskID, newStatus string) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		err := updateTaskStatusViaMCP(ctx, server, taskID, newStatus)
		return statusUpdateDoneMsg{err: err}
	}
}

// bulkUpdateStatusCmd updates multiple tasks' status via task_workflow MCP tool and returns bulkStatusUpdateDoneMsg.
func bulkUpdateStatusCmd(server framework.MCPServer, taskIDs []string, newStatus string) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		updated, err := bulkUpdateStatusViaMCP(ctx, server, taskIDs, newStatus)
		return bulkStatusUpdateDoneMsg{updated: updated, total: len(taskIDs), err: err}
	}
}

func loadScorecard(projectRoot string, fullMode bool) tea.Cmd {
	return func() tea.Msg {
		if projectRoot == "" {
			return scorecardLoadedMsg{err: fmt.Errorf("no project root")}
		}

		ctx := context.Background()

		var combined strings.Builder

		var recommendations []string

		if tools.IsGoProject() {
			opts := &tools.ScorecardOptions{FastMode: !fullMode}

			scorecard, err := tools.GenerateGoScorecard(ctx, projectRoot, opts)
			if err != nil {
				return scorecardLoadedMsg{err: err}
			}

			combined.WriteString("=== Go Scorecard ===\n\n")
			combined.WriteString(tools.FormatGoScorecard(scorecard))
			recommendations = scorecard.Recommendations
		}

		overviewText, err := tools.GetOverviewText(ctx, projectRoot)
		if err != nil {
			if combined.Len() > 0 {
				combined.WriteString("\n\n=== Project Overview ===\n\n(overview failed: ")
				combined.WriteString(err.Error())
				combined.WriteString(")")
			} else {
				return scorecardLoadedMsg{err: err}
			}
		} else {
			if combined.Len() > 0 {
				combined.WriteString("\n\n")
			}

			combined.WriteString("=== Project Overview ===\n\n")
			combined.WriteString(overviewText)
		}

		return scorecardLoadedMsg{text: combined.String(), recommendations: recommendations}
	}
}

func loadHandoffs(server framework.MCPServer) tea.Cmd {
	return func() tea.Msg {
		if server == nil {
			return handoffLoadedMsg{err: fmt.Errorf("no server")}
		}

		ctx := context.Background()
		args := map[string]interface{}{
			"action":         "handoff",
			"sub_action":     "list",
			"limit":          float64(20),
			"include_closed": false,
		}

		argsBytes, err := json.Marshal(args)
		if err != nil {
			return handoffLoadedMsg{err: err}
		}

		result, err := server.CallTool(ctx, "session", argsBytes)
		if err != nil {
			return handoffLoadedMsg{err: err}
		}

		var b strings.Builder
		for _, c := range result {
			b.WriteString(c.Text)
			b.WriteString("\n")
		}

		text := strings.TrimSpace(b.String())

		var entries []map[string]interface{}

		var payload struct {
			Handoffs []map[string]interface{} `json:"handoffs"`
		}

		if err := json.Unmarshal([]byte(text), &payload); err == nil && len(payload.Handoffs) > 0 {
			entries = payload.Handoffs
		}

		return handoffLoadedMsg{text: text, entries: entries, err: nil}
	}
}

func runWavesRefreshTools(server framework.MCPServer) tea.Cmd {
	return func() tea.Msg {
		if server == nil {
			return wavesRefreshDoneMsg{err: fmt.Errorf("no server")}
		}

		ctx := context.Background()

		syncArgs, _ := json.Marshal(map[string]interface{}{"action": "sync", "sub_action": "list"})
		if _, err := server.CallTool(ctx, "task_workflow", syncArgs); err != nil {
			return wavesRefreshDoneMsg{err: fmt.Errorf("task_workflow sync: %w", err)}
		}

		taArgs, _ := json.Marshal(map[string]interface{}{"action": "parallelization", "output_format": "text"})
		if _, err := server.CallTool(ctx, "task_analysis", taArgs); err != nil {
			return wavesRefreshDoneMsg{err: fmt.Errorf("task_analysis parallelization: %w", err)}
		}

		return wavesRefreshDoneMsg{err: nil}
	}
}

func runTaskAnalysis(server framework.MCPServer, action string) tea.Cmd {
	return func() tea.Msg {
		if server == nil {
			return taskAnalysisLoadedMsg{action: action, err: fmt.Errorf("no server")}
		}

		ctx := context.Background()
		args, _ := json.Marshal(map[string]interface{}{"action": action, "output_format": "text"})

		result, err := server.CallTool(ctx, "task_analysis", args)
		if err != nil {
			return taskAnalysisLoadedMsg{action: action, err: err}
		}

		var b strings.Builder
		for _, c := range result {
			b.WriteString(c.Text)
			b.WriteString("\n")
		}

		return taskAnalysisLoadedMsg{text: strings.TrimSpace(b.String()), action: action, err: nil}
	}
}

func runReportUpdateWavesFromPlan(server framework.MCPServer, projectRoot string) tea.Cmd {
	return func() tea.Msg {
		if server == nil {
			return updateWavesFromPlanDoneMsg{message: "", err: fmt.Errorf("no server")}
		}

		ctx := context.Background()
		args, _ := json.Marshal(map[string]interface{}{
			"action": "update_waves_from_plan",
		})

		result, err := server.CallTool(ctx, "report", args)
		if err != nil {
			return updateWavesFromPlanDoneMsg{message: "", err: err}
		}

		var msg string
		for _, c := range result {
			msg += c.Text
		}

		return updateWavesFromPlanDoneMsg{message: strings.TrimSpace(msg), err: nil}
	}
}

func runEnqueueWave(projectRoot string, waveLevel int) tea.Cmd {
	return func() tea.Msg {
		cfg := queue.ConfigFromEnv()
		if !cfg.Enabled() {
			return enqueueWaveDoneMsg{waveLevel: waveLevel, err: fmt.Errorf("REDIS_ADDR not set")}
		}
		producer, err := queue.NewProducer(cfg)
		if err != nil {
			return enqueueWaveDoneMsg{waveLevel: waveLevel, err: err}
		}
		defer producer.Close()

		enqueued, err := producer.EnqueueWave(context.Background(), projectRoot, waveLevel)
		return enqueueWaveDoneMsg{waveLevel: waveLevel, enqueued: enqueued, err: err}
	}
}

func runReportParallelExecutionPlan(server framework.MCPServer, projectRoot string) tea.Cmd {
	return func() tea.Msg {
		if server == nil {
			return taskAnalysisApproveDoneMsg{err: fmt.Errorf("no server")}
		}

		ctx := context.Background()
		outputPath := filepath.Join(projectRoot, ".cursor", "plans", "parallel-execution-subagents.plan.md")
		args, _ := json.Marshal(map[string]interface{}{
			"action":        "execution_plan",
			"output_format": "subagents_plan",
			"output_path":   outputPath,
		})

		result, err := server.CallTool(ctx, "task_analysis", args)
		if err != nil {
			return taskAnalysisApproveDoneMsg{err: err}
		}

		var msg string
		for _, c := range result {
			msg += c.Text
		}

		msg = strings.TrimSpace(msg)
		if msg == "" {
			msg = "Waves plan written to .cursor/plans/parallel-execution-subagents.plan.md"
		}

		return taskAnalysisApproveDoneMsg{message: msg, err: nil}
	}
}

func moveTaskToWaveCmd(server framework.MCPServer, taskID string, newDeps []string) tea.Cmd {
	return func() tea.Msg {
		if taskID == "" {
			return moveTaskToWaveDoneMsg{taskID: "", err: fmt.Errorf("no task")}
		}

		ctx := context.Background()
		err := moveTaskToWaveViaMCP(ctx, server, taskID, newDeps)
		return moveTaskToWaveDoneMsg{taskID: taskID, err: err}
	}
}

func runHandoffAction(server framework.MCPServer, projectRoot string, handoffIDs []string, action string) tea.Cmd {
	return func() tea.Msg {
		if server == nil || len(handoffIDs) == 0 {
			return handoffActionDoneMsg{action: action, err: fmt.Errorf("no server or no handoff IDs")}
		}

		ctx := context.Background()
		args := map[string]interface{}{
			"action":      "handoff",
			"sub_action":  action,
			"handoff_ids": handoffIDs,
		}

		argsBytes, err := json.Marshal(args)
		if err != nil {
			return handoffActionDoneMsg{action: action, err: err}
		}

		_, err = server.CallTool(ctx, "session", argsBytes)
		if err != nil {
			return handoffActionDoneMsg{action: action, err: err}
		}

		return handoffActionDoneMsg{action: action, updated: len(handoffIDs)}
	}
}

func recommendationToCommand(rec string) (name string, args []string, ok bool) {
	rec = strings.TrimSpace(rec)

	switch {
	case strings.Contains(rec, "go mod tidy"):
		return "make", []string{"go-mod-tidy"}, true
	case strings.Contains(rec, "go fmt"):
		return "make", []string{"go-fmt"}, true
	case strings.Contains(rec, "go vet"):
		return "make", []string{"go-vet"}, true
	case strings.Contains(rec, "Fix Go build"):
		return "make", []string{"build"}, true
	case strings.Contains(rec, "golangci-lint issues"):
		return "make", []string{"golangci-lint-fix"}, true
	case strings.Contains(rec, "failing Go tests"), strings.Contains(rec, "go test"):
		return "make", []string{"test"}, true
	case strings.Contains(rec, "govulncheck"):
		return "make", []string{"govulncheck"}, true
	case strings.Contains(rec, "test coverage"):
		return "make", []string{"test-coverage"}, true
	default:
		return "", nil, false
	}
}

func runRecommendationCmd(projectRoot, rec string) tea.Cmd {
	return func() tea.Msg {
		name, args, ok := recommendationToCommand(rec)
		if !ok {
			return runRecommendationResultMsg{output: "", err: fmt.Errorf("no runnable command for this recommendation")}
		}

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()

		cmd := exec.CommandContext(ctx, name, args...)
		cmd.Dir = projectRoot

		out, err := cmd.CombinedOutput()
		if err != nil {
			return runRecommendationResultMsg{output: string(out), err: err}
		}

		return runRecommendationResultMsg{output: string(out), err: nil}
	}
}

func showConfigSection(section configSection, cfg *config.FullConfig) tea.Cmd {
	return func() tea.Msg {
		var details strings.Builder

		details.WriteString(fmt.Sprintf("\nConfig Section: %s\n", section.name))
		details.WriteString(fmt.Sprintf("Description: %s\n", section.description))
		details.WriteString("\nKey values:\n")

		switch section.name {
		case "Timeouts":
			details.WriteString(fmt.Sprintf("  Task Lock Lease: %v\n", cfg.Timeouts.TaskLockLease))
			details.WriteString(fmt.Sprintf("  Tool Default: %v\n", cfg.Timeouts.ToolDefault))
			details.WriteString(fmt.Sprintf("  HTTP Client: %v\n", cfg.Timeouts.HTTPClient))
		case "Thresholds":
			details.WriteString(fmt.Sprintf("  Similarity Threshold: %.2f\n", cfg.Thresholds.SimilarityThreshold))
			details.WriteString(fmt.Sprintf("  Min Coverage: %d%%\n", cfg.Thresholds.MinCoverage))
			details.WriteString(fmt.Sprintf("  Min Task Confidence: %.2f\n", cfg.Thresholds.MinTaskConfidence))
		case "Tasks":
			details.WriteString(fmt.Sprintf("  Default Status: %s\n", cfg.Tasks.DefaultStatus))
			details.WriteString(fmt.Sprintf("  Default Priority: %s\n", cfg.Tasks.DefaultPriority))
			details.WriteString(fmt.Sprintf("  Stale Threshold: %d hours\n", cfg.Tasks.StaleThresholdHours))
		case "Database":
			details.WriteString(fmt.Sprintf("  SQLite Path: %s\n", cfg.Database.SQLitePath))
			details.WriteString(fmt.Sprintf("  Max Connections: %d\n", cfg.Database.MaxConnections))
			details.WriteString(fmt.Sprintf("  Query Timeout: %v\n", cfg.Database.QueryTimeout))
		case "Security":
			details.WriteString(fmt.Sprintf("  Rate Limit Enabled: %v\n", cfg.Security.RateLimit.Enabled))
			details.WriteString(fmt.Sprintf("  Max File Size: %d bytes\n", cfg.Thresholds.MaxFileSize))
			details.WriteString(fmt.Sprintf("  Max Path Depth: %d\n", cfg.Thresholds.MaxPathDepth))
		}

		details.WriteString("\nUpdate from TUI: press ")
		details.WriteString("u")
		details.WriteString(" or ")
		details.WriteString("s")
		details.WriteString(" to save to .exarp/config.pb (protobuf).")
		details.WriteString("\nOr use CLI: 'exarp-go config export yaml' to edit as YAML, then 'convert yaml protobuf' to save.")

		return configSectionDetailMsg{text: details.String()}
	}
}

func saveConfig(projectRoot string, cfg *config.FullConfig) tea.Cmd {
	return func() tea.Msg {
		if err := config.WriteConfigToProtobufFile(projectRoot, cfg); err != nil {
			return configSaveResultMsg{message: fmt.Sprintf("Save failed: %v", err), success: false}
		}
		return configSaveResultMsg{message: "Config saved to .exarp/config.pb", success: true}
	}
}

func runChildAgentCmd(projectRoot, prompt string, kind ChildAgentKind) tea.Cmd {
	return func() tea.Msg {
		r, done := RunChildAgentWithOutputCapture(projectRoot, prompt)
		r.Kind = kind

		return tea.Batch(
			func() tea.Msg { return childAgentResultMsg{Result: r} },
			func() tea.Msg {
				out := <-done
				return jobCompletedMsg(out)
			},
		)
	}
}

func runChildAgentCmdInteractive(projectRoot, prompt string, kind ChildAgentKind) tea.Cmd {
	return func() tea.Msg {
		r := RunChildAgentInteractive(projectRoot, prompt)
		r.Kind = kind

		return childAgentResultMsg{Result: r}
	}
}
