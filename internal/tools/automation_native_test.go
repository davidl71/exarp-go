package tools

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/davidl71/exarp-go/internal/database"
)

// automationTestTimeout is the max duration per automation subtest. Prevents
// long-running tests (e.g. discover with Apple FM, sprint with many subtasks) from blocking the suite.
// 25s keeps failures fast vs package default 90s.
const automationTestTimeout = 25 * time.Second

// runWithTimeout runs fn in a goroutine and fails the test if it exceeds d.
// Use for subtests that may block on I/O or Apple FM so they fail fast instead of waiting for package timeout.
func runWithTimeout(t *testing.T, d time.Duration, fn func()) {
	t.Helper()

	done := make(chan struct{}, 1)

	go func() {
		defer func() { done <- struct{}{} }()

		fn()
	}()

	select {
	case <-done:
		return
	case <-time.After(d):
		t.Fatalf("timeout: test exceeded %v (use -short to skip long-running tests)", d)
	}
}

func TestHandleAutomationNative(t *testing.T) {
	tests := []struct {
		name      string
		action    string
		params    map[string]interface{}
		wantError bool
		longRun   bool // longRun tests are skipped when -short is set
	}{
		{
			name:      "daily action",
			action:    "daily",
			params:    map[string]interface{}{"action": "daily"},
			wantError: false,
		},
		{
			name:      "execution_cockpit action",
			action:    "execution_cockpit",
			params:    map[string]interface{}{"action": "execution_cockpit", "output_format": "json"},
			wantError: false,
		},
		{
			name:      "discover action",
			action:    "discover",
			params:    map[string]interface{}{"action": "discover", "use_llm": false},
			wantError: false,
			longRun:   true, // discover can block on Apple FM when CGO=1
		},
		{
			name:      "nightly action",
			action:    "nightly",
			params:    map[string]interface{}{"action": "nightly"},
			wantError: false,
		},
		{
			name:      "sprint action",
			action:    "sprint",
			params:    map[string]interface{}{"action": "sprint"},
			wantError: false,
			longRun:   true, // sprint runs many subtasks and can exceed quick timeout
		},
		{
			name:      "unknown action",
			action:    "unknown",
			params:    map[string]interface{}{"action": "unknown"},
			wantError: true,
		},
		{
			name:      "empty action defaults to daily",
			action:    "",
			params:    map[string]interface{}{},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.longRun && testing.Short() {
				t.Skip("skipping long-running discover in short mode")
			}

			runWithTimeout(t, automationTestTimeout, func() {
				ctx := context.Background()

				result, err := handleAutomationNative(ctx, tt.params)
				if (err != nil) != tt.wantError {
					t.Errorf("handleAutomationNative() error = %v, wantError %v", err, tt.wantError)
					return
				}

				if !tt.wantError {
					if result == nil || len(result) == 0 {
						t.Error("handleAutomationNative() returned empty result")
						return
					}
					// Verify result is valid JSON
					var resultData map[string]interface{}
					if err := json.Unmarshal([]byte(result[0].Text), &resultData); err != nil {
						t.Errorf("handleAutomationNative() returned invalid JSON: %v", err)
					}
				}
			})
		})
	}
}

func TestHandleAutomationDaily(t *testing.T) {
	tests := []struct {
		name      string
		params    map[string]interface{}
		wantError bool
	}{
		{
			name:      "valid daily action",
			params:    map[string]interface{}{"action": "daily"},
			wantError: false,
		},
		{
			name:      "empty params",
			params:    map[string]interface{}{},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			result, err := handleAutomationDaily(ctx, tt.params)
			if (err != nil) != tt.wantError {
				t.Errorf("handleAutomationDaily() error = %v, wantError %v", err, tt.wantError)
				return
			}

			if !tt.wantError && (result == nil || len(result) == 0) {
				t.Error("handleAutomationDaily() returned empty result")
			}
		})
	}
}

func TestHandleAutomationNightly(t *testing.T) {
	tests := []struct {
		name      string
		params    map[string]interface{}
		wantError bool
	}{
		{
			name: "valid nightly action",
			params: map[string]interface{}{
				"action":             "nightly",
				"max_tasks_per_host": 5,
				"max_parallel_tasks": 10,
				"dry_run":            true,
			},
			wantError: false,
		},
		{
			name:      "empty params",
			params:    map[string]interface{}{"action": "nightly"},
			wantError: false,
		},
		{
			name: "with priority filter",
			params: map[string]interface{}{
				"action":          "nightly",
				"priority_filter": "high",
				"dry_run":         true,
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			result, err := handleAutomationNightly(ctx, tt.params)
			if (err != nil) != tt.wantError {
				t.Errorf("handleAutomationNightly() error = %v, wantError %v", err, tt.wantError)
				return
			}

			if !tt.wantError && (result == nil || len(result) == 0) {
				t.Error("handleAutomationNightly() returned empty result")
			}
		})
	}
}

func TestHandleAutomationSprint(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping sprint test in short mode (can reach real FM-backed hierarchy analysis)")
	}

	tests := []struct {
		name      string
		params    map[string]interface{}
		wantError bool
	}{
		{
			name: "valid sprint action",
			params: map[string]interface{}{
				"action":             "sprint",
				"max_iterations":     1,
				"auto_approve":       false,
				"extract_subtasks":   false,
				"run_analysis_tools": false,
				"run_testing_tools":  false,
				"dry_run":            true,
			},
			wantError: false,
		},
		{
			name:      "empty params",
			params:    map[string]interface{}{"action": "sprint"},
			wantError: false,
		},
		{
			name: "sprint with all features enabled",
			params: map[string]interface{}{
				"action":             "sprint",
				"max_iterations":     2,
				"auto_approve":       true,
				"extract_subtasks":   true,
				"run_analysis_tools": true,
				"run_testing_tools":  true,
				"dry_run":            true,
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			result, err := handleAutomationSprint(ctx, tt.params)
			if (err != nil) != tt.wantError {
				t.Errorf("handleAutomationSprint() error = %v, wantError %v", err, tt.wantError)
				return
			}

			if !tt.wantError && (result == nil || len(result) == 0) {
				t.Error("handleAutomationSprint() returned empty result")
				return
			}

			if !tt.wantError {
				// Verify result contains expected fields (action is under results)
				var resultData map[string]interface{}
				if err := json.Unmarshal([]byte(result[0].Text), &resultData); err == nil {
					if results, ok := resultData["results"].(map[string]interface{}); ok {
						if action, ok := results["action"].(string); !ok || action != "sprint" {
							t.Errorf("handleAutomationSprint() result results.action = %v, want 'sprint'", action)
						}
					} else if action, ok := resultData["action"].(string); !ok || action != "sprint" {
						t.Errorf("handleAutomationSprint() result action = %v, want 'sprint'", action)
					}
				}
			}
		})
	}
}

func TestHandleAutomationDiscover(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping discover test in short mode (can block on Apple FM when CGO=1)")
	}

	tests := []struct {
		name      string
		params    map[string]interface{}
		wantError bool
	}{
		{
			name:      "valid discover action",
			params:    map[string]interface{}{"action": "discover", "use_llm": false},
			wantError: false,
		},
		{
			name:      "empty params",
			params:    map[string]interface{}{"use_llm": false},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runWithTimeout(t, automationTestTimeout, func() {
				ctx := context.Background()

				result, err := handleAutomationDiscover(ctx, tt.params)
				if (err != nil) != tt.wantError {
					t.Errorf("handleAutomationDiscover() error = %v, wantError %v", err, tt.wantError)
					return
				}

				if !tt.wantError && (result == nil || len(result) == 0) {
					t.Error("handleAutomationDiscover() returned empty result")
				}
			})
		})
	}
}

func TestHandleAutomationScheduleDryRun(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	runWithTimeout(t, automationTestTimeout, func() {
		ctx := context.Background()

		result, err := handleAutomationNative(ctx, map[string]interface{}{
			"action":           "schedule",
			"target_action":    "daily",
			"schedule_label":   "test-schedule",
			"interval_seconds": 120,
			"dry_run":          true,
		})
		if err != nil {
			t.Fatalf("handleAutomationNative(schedule) error = %v", err)
		}
		if result == nil || len(result) == 0 {
			t.Fatal("handleAutomationNative(schedule) returned empty result")
		}

		var payload map[string]interface{}
		if err := json.Unmarshal([]byte(result[0].Text), &payload); err != nil {
			t.Fatalf("schedule result JSON invalid: %v", err)
		}
		if got := payload["action"]; got != "schedule" {
			t.Fatalf("action = %#v, want schedule", got)
		}
		if got := payload["status"]; got != "success" {
			t.Fatalf("status = %#v, want success", got)
		}

		results, ok := payload["results"].(map[string]interface{})
		if !ok {
			t.Fatalf("results missing or wrong type: %#v", payload["results"])
		}
		if got := results["target_action"]; got != "daily" {
			t.Fatalf("results.target_action = %#v, want daily", got)
		}
		if got := results["schedule_label"]; got != "test-schedule" {
			t.Fatalf("results.schedule_label = %#v, want test-schedule", got)
		}
		artifacts, ok := results["artifacts"].(map[string]interface{})
		if !ok {
			t.Fatalf("results.artifacts missing or wrong type: %#v", results["artifacts"])
		}
		if _, ok := artifacts["script"].(string); !ok {
			t.Fatalf("artifacts.script missing or wrong type: %#v", artifacts["script"])
		}
	})
}

func TestHandleAutomationScheduleDryRunExecutionCockpit(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	runWithTimeout(t, automationTestTimeout, func() {
		ctx := context.Background()

		result, err := handleAutomationNative(ctx, map[string]interface{}{
			"action":           "schedule",
			"target_action":    "execution_cockpit",
			"schedule_label":   "test-schedule-cockpit",
			"interval_seconds": 120,
			"dry_run":          true,
		})
		if err != nil {
			t.Fatalf("handleAutomationNative(schedule cockpit) error = %v", err)
		}
		if result == nil || len(result) == 0 {
			t.Fatal("handleAutomationNative(schedule cockpit) returned empty result")
		}

		var payload map[string]interface{}
		if err := json.Unmarshal([]byte(result[0].Text), &payload); err != nil {
			t.Fatalf("schedule cockpit result JSON invalid: %v", err)
		}
		if got := payload["action"]; got != "schedule" {
			t.Fatalf("action = %#v, want schedule", got)
		}
		if got := payload["status"]; got != "success" {
			t.Fatalf("status = %#v, want success", got)
		}

		results, ok := payload["results"].(map[string]interface{})
		if !ok {
			t.Fatalf("results missing or wrong type: %#v", payload["results"])
		}
		if got := results["target_action"]; got != "execution_cockpit" {
			t.Fatalf("results.target_action = %#v, want execution_cockpit", got)
		}
		if got := results["schedule_label"]; got != "test-schedule-cockpit" {
			t.Fatalf("results.schedule_label = %#v, want test-schedule-cockpit", got)
		}
	})
}

func TestAutomationScheduleRenderers(t *testing.T) {
	cfg := &automationScheduleConfig{
		TargetAction:    "daily",
		ScheduleLabel:   "com.davidl71.exarpgo.automation.daily",
		IntervalSeconds: 600,
		RunAtLoad:       true,
		Enabled:         true,
		ProjectRoot:     "/tmp/exarp-go",
		BinaryPath:      "/usr/local/bin/exarp-go",
		HomeDir:         t.TempDir(),
	}
	cfg.preparePaths()

	artifacts, err := buildAutomationScheduleArtifacts(cfg)
	if err != nil {
		t.Fatalf("buildAutomationScheduleArtifacts() error = %v", err)
	}
	if !strings.Contains(artifacts.ScriptContents, "-tool automation") {
		t.Fatalf("script contents missing automation invocation: %s", artifacts.ScriptContents)
	}

	plist, err := renderLaunchdPlist(cfg)
	if err != nil {
		t.Fatalf("renderLaunchdPlist() error = %v", err)
	}
	if !strings.Contains(plist, "<key>StartInterval</key>") || !strings.Contains(plist, "<integer>600</integer>") {
		t.Fatalf("launchd plist missing start interval: %s", plist)
	}
	if !strings.Contains(plist, cfg.ScriptPath) {
		t.Fatalf("launchd plist missing script path: %s", plist)
	}

	service, timer := renderSystemdUnit(cfg)
	if !strings.Contains(service, "ExecStart="+strings.ReplaceAll(cfg.ScriptPath, " ", `\x20`)) {
		t.Fatalf("systemd service missing script exec: %s", service)
	}
	if !strings.Contains(timer, "OnUnitActiveSec=600sec") {
		t.Fatalf("systemd timer missing interval: %s", timer)
	}
}

func TestHandleAutomationUnscheduleDryRun(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	runWithTimeout(t, automationTestTimeout, func() {
		ctx := context.Background()

		result, err := handleAutomationNative(ctx, map[string]interface{}{
			"action":         "unschedule",
			"target_action":  "daily",
			"schedule_label": "test-schedule",
			"dry_run":        true,
		})
		if err != nil {
			t.Fatalf("handleAutomationNative(unschedule) error = %v", err)
		}
		if result == nil || len(result) == 0 {
			t.Fatal("handleAutomationNative(unschedule) returned empty result")
		}

		var payload map[string]interface{}
		if err := json.Unmarshal([]byte(result[0].Text), &payload); err != nil {
			t.Fatalf("unschedule result JSON invalid: %v", err)
		}
		if got := payload["action"]; got != "unschedule" {
			t.Fatalf("action = %#v, want unschedule", got)
		}
		results, ok := payload["results"].(map[string]interface{})
		if !ok {
			t.Fatalf("results missing or wrong type: %#v", payload["results"])
		}
		if got := results["preview"]; got != true {
			t.Fatalf("results.preview = %#v, want true", got)
		}
	})
}

func TestBeginAutomationRunSkipsActiveRun(t *testing.T) {
	_, self, _, _ := runtime.Caller(0)
	repoRoot := filepath.Dir(filepath.Dir(filepath.Dir(self)))
	tmpDir := t.TempDir()

	cfg, err := database.LoadConfig(tmpDir)
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}
	cfg.Driver = database.DriverSQLite
	cfg.DSN = filepath.Join(tmpDir, ".todo2", "todo2.db")
	cfg.MigrationsDir = filepath.Join(repoRoot, "migrations")
	cfg.AutoMigrate = true

	if err := database.InitWithConfig(cfg); err != nil {
		t.Fatalf("InitWithConfig() error = %v", err)
	}
	t.Cleanup(func() { _ = database.Close() })

	db, err := database.GetDBx()
	if err != nil {
		t.Fatalf("GetDBx() error = %v", err)
	}

	now := time.Now().Unix()
	pid := os.Getpid()
	_, err = db.ExecContext(context.Background(), `
		INSERT INTO automation_runs (
			schedule_label, action, pid, host, status, started_at, created_at, updated_at
		) VALUES (?, ?, ?, ?, 'running', ?, ?, ?)
	`, "test-schedule", "daily", pid, "test-host", now, now, now)
	if err != nil {
		t.Fatalf("seed automation_runs row error = %v", err)
	}

	guard, skipResult, err := beginAutomationRun(context.Background(), "daily", "test-schedule")
	if err != nil {
		t.Fatalf("beginAutomationRun() error = %v", err)
	}
	if guard != nil {
		t.Fatalf("beginAutomationRun() guard = %#v, want nil", guard)
	}
	if skipResult == nil {
		t.Fatal("beginAutomationRun() skipResult = nil, want skip details")
	}
	if got := skipResult["status"]; got != "skipped" {
		t.Fatalf("skipResult.status = %#v, want skipped", got)
	}
}
