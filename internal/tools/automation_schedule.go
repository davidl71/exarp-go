// automation_schedule.go — Automation schedule installer and run guards.
package tools

import (
	"context"
	"database/sql"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/davidl71/exarp-go/internal/database"
	"github.com/davidl71/exarp-go/internal/framework"
	"github.com/davidl71/exarp-go/proto"
	"github.com/spf13/cast"
)

const (
	defaultAutomationLabelPrefix   = "com.davidl71.exarpgo.automation"
	defaultAutomationIntervalMins  = 24 * 60
	automationScriptDirNameDarwin  = "Library/Application Support/exarp-go/automation"
	automationScriptDirNameLinux   = ".local/share/exarp-go/automation"
	automationSystemdDirNameLinux  = ".config/systemd/user"
	automationLaunchdDirNameDarwin = "Library/LaunchAgents"
)

type automationScheduleConfig struct {
	TargetAction    string
	ScheduleLabel   string
	IntervalSeconds int
	RunAtLoad       bool
	Enabled         bool
	DryRun          bool
	ProjectRoot     string
	BinaryPath      string
	HomeDir         string
	ScriptPath      string
	PlistPath       string
	ServicePath     string
	TimerPath       string
	LogPath         string
}

type automationRunGuard struct {
	RunID         int64
	ScheduleLabel string
	Action        string
	PID           int
	StartedAt     time.Time
}

type automationRunRow struct {
	ID            int64          `db:"id"`
	ScheduleLabel string         `db:"schedule_label"`
	Action        string         `db:"action"`
	PID           int            `db:"pid"`
	Status        string         `db:"status"`
	StartedAt     int64          `db:"started_at"`
	EndedAt       sql.NullInt64  `db:"ended_at"`
	ErrorText     sql.NullString `db:"error_text"`
}

func handleAutomationSchedule(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	cfg, err := resolveAutomationScheduleConfig(params)
	if err != nil {
		return nil, err
	}

	result := map[string]interface{}{
		"action":           "schedule",
		"target_action":    cfg.TargetAction,
		"schedule_label":   cfg.ScheduleLabel,
		"interval_seconds": cfg.IntervalSeconds,
		"run_at_load":      cfg.RunAtLoad,
		"enabled":          cfg.Enabled,
		"dry_run":          cfg.DryRun,
		"platform":         runtime.GOOS,
	}

	if projectRoot, err := FindProjectRoot(); err == nil {
		if taskOverlaps, fileConflicts, forbidden, errDetect := DetectConflicts(ctx, projectRoot); errDetect == nil {
			if len(taskOverlaps) > 0 || len(fileConflicts) > 0 || len(forbidden) > 0 {
				result["conflicts"] = map[string]interface{}{
					"task_overlap": taskOverlaps,
					"file":         fileConflicts,
					"forbidden":    forbidden,
				}
			}
		}
	}

	artifacts, err := buildAutomationScheduleArtifacts(cfg)
	if err != nil {
		return nil, err
	}

	result["artifacts"] = artifacts.toMap()
	result["paths"] = map[string]interface{}{
		"script":  cfg.ScriptPath,
		"plist":   cfg.PlistPath,
		"service": cfg.ServicePath,
		"timer":   cfg.TimerPath,
		"log":     cfg.LogPath,
	}

	if cfg.DryRun {
		result["status"] = "success"
		result["installed"] = false
		result["loaded"] = false
		return framework.FormatResult(scheduleResponseToMap("schedule", result), "")
	}

	if err := writeAutomationScheduleArtifacts(cfg, artifacts); err != nil {
		return nil, err
	}

	if !cfg.Enabled {
		result["status"] = "success"
		result["installed"] = true
		result["loaded"] = false
		result["enabled"] = false
		result["note"] = "schedule written but left disabled"
		return framework.FormatResult(scheduleResponseToMap("schedule", result), "")
	}

	loaded := false
	switch runtime.GOOS {
	case "darwin":
		loaded, err = loadLaunchdJob(cfg)
	case "linux":
		loaded, err = loadSystemdTimer(cfg)
	default:
		result["status"] = "unsupported"
		result["installed"] = true
		result["loaded"] = false
		result["fallback"] = "schedule files written; start them manually on this platform"
		return framework.FormatResult(scheduleResponseToMap("schedule", result), "")
	}
	if err != nil {
		return nil, err
	}

	result["status"] = "success"
	result["installed"] = true
	result["loaded"] = loaded
	return framework.FormatResult(scheduleResponseToMap("schedule", result), "")
}

func handleAutomationUnschedule(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	cfg, err := resolveAutomationScheduleConfig(params)
	if err != nil {
		return nil, err
	}

	result := map[string]interface{}{
		"action":         "unschedule",
		"target_action":  cfg.TargetAction,
		"schedule_label": cfg.ScheduleLabel,
		"platform":       runtime.GOOS,
		"dry_run":        cfg.DryRun,
		"paths": map[string]interface{}{
			"script":  cfg.ScriptPath,
			"plist":   cfg.PlistPath,
			"service": cfg.ServicePath,
			"timer":   cfg.TimerPath,
		},
	}

	if cfg.DryRun {
		result["status"] = "success"
		result["removed"] = false
		result["unloaded"] = false
		result["preview"] = true
		return framework.FormatResult(scheduleResponseToMap("unschedule", result), "")
	}

	unloaded := false
	switch runtime.GOOS {
	case "darwin":
		unloaded, err = unloadLaunchdJob(cfg)
	case "linux":
		unloaded, err = unloadSystemdTimer(cfg)
	default:
		result["status"] = "unsupported"
		result["removed"] = false
		result["unloaded"] = false
		return framework.FormatResult(scheduleResponseToMap("unschedule", result), "")
	}
	if err != nil {
		return nil, err
	}

	removed, removeErr := removeAutomationScheduleArtifacts(cfg)
	if removeErr != nil {
		return nil, removeErr
	}

	result["status"] = "success"
	result["removed"] = removed
	result["unloaded"] = unloaded
	return framework.FormatResult(scheduleResponseToMap("unschedule", result), "")
}

func resolveAutomationScheduleConfig(params map[string]interface{}) (*automationScheduleConfig, error) {
	targetAction := strings.TrimSpace(cast.ToString(params["target_action"]))
	if targetAction == "" {
		targetAction = strings.TrimSpace(cast.ToString(params["schedule_action"]))
	}
	if targetAction == "" {
		targetAction = "daily"
	}

	if !isAutomationTargetAction(targetAction) {
		return nil, fmt.Errorf("unknown target_action: %s (use daily, nightly, sprint, discover, or execution_cockpit)", targetAction)
	}

	label := strings.TrimSpace(cast.ToString(params["schedule_label"]))
	if label == "" {
		label = defaultAutomationScheduleLabel(targetAction)
	}
	label = sanitizeAutomationLabel(label)

	intervalSeconds := cast.ToInt(params["interval_seconds"])
	if intervalSeconds <= 0 {
		intervalSeconds = cast.ToInt(params["interval_minutes"]) * 60
	}
	if intervalSeconds <= 0 {
		intervalSeconds = defaultAutomationIntervalMins * 60
	}

	runAtLoad := cast.ToBool(params["run_at_load"])
	enabled := true
	if val, ok := params["enabled"]; ok {
		enabled = cast.ToBool(val)
	}

	projectRoot := strings.TrimSpace(cast.ToString(params["project_root"]))
	if projectRoot == "" {
		if root, err := FindProjectRoot(); err == nil && root != "" {
			projectRoot = root
		} else if wd, wdErr := os.Getwd(); wdErr == nil {
			projectRoot = wd
		} else {
			projectRoot = "."
		}
	}

	homeDir, err := os.UserHomeDir()
	if err != nil || homeDir == "" {
		homeDir = projectRoot
	}

	binaryPath, err := os.Executable()
	if err != nil || binaryPath == "" {
		return nil, fmt.Errorf("failed to resolve exarp-go executable: %w", err)
	}
	if absBinary, absErr := filepath.Abs(binaryPath); absErr == nil {
		binaryPath = absBinary
	}

	cfg := &automationScheduleConfig{
		TargetAction:    targetAction,
		ScheduleLabel:   label,
		IntervalSeconds: intervalSeconds,
		RunAtLoad:       runAtLoad,
		Enabled:         enabled,
		DryRun:          cast.ToBool(params["dry_run"]),
		ProjectRoot:     projectRoot,
		BinaryPath:      binaryPath,
		HomeDir:         homeDir,
	}
	cfg.preparePaths()
	return cfg, nil
}

func (cfg *automationScheduleConfig) preparePaths() {
	scriptRoot := filepath.Join(cfg.HomeDir, automationScriptDirForGOOS())
	cfg.ScriptPath = filepath.Join(scriptRoot, cfg.ScheduleLabel+".sh")
	cfg.PlistPath = filepath.Join(filepath.Join(cfg.HomeDir, automationLaunchdDirNameDarwin), cfg.ScheduleLabel+".plist")
	cfg.ServicePath = filepath.Join(filepath.Join(cfg.HomeDir, automationSystemdDirNameLinux), cfg.ScheduleLabel+".service")
	cfg.TimerPath = filepath.Join(filepath.Join(cfg.HomeDir, automationSystemdDirNameLinux), cfg.ScheduleLabel+".timer")
	cfg.LogPath = filepath.Join(cfg.HomeDir, ".local/state/exarp-go/automation", cfg.ScheduleLabel+".log")
	_ = os.MkdirAll(scriptRoot, 0o755)
	_ = os.MkdirAll(filepath.Dir(cfg.PlistPath), 0o755)
	_ = os.MkdirAll(filepath.Dir(cfg.ServicePath), 0o755)
	_ = os.MkdirAll(filepath.Dir(cfg.LogPath), 0o755)
}

func automationScriptDirForGOOS() string {
	if runtime.GOOS == "darwin" {
		return automationScriptDirNameDarwin
	}
	return automationScriptDirNameLinux
}

type automationScheduleArtifacts struct {
	ScriptContents  string
	PlistContents   string
	ServiceContents string
	TimerContents   string
}

func (a automationScheduleArtifacts) toMap() map[string]interface{} {
	return map[string]interface{}{
		"script":  a.ScriptContents,
		"plist":   a.PlistContents,
		"service": a.ServiceContents,
		"timer":   a.TimerContents,
	}
}

func buildAutomationScheduleArtifacts(cfg *automationScheduleConfig) (*automationScheduleArtifacts, error) {
	args, err := json.Marshal(map[string]interface{}{
		"action":         cfg.TargetAction,
		"schedule_label": cfg.ScheduleLabel,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal schedule args: %w", err)
	}

	scriptContents := fmt.Sprintf(`#!/bin/sh
set -eu
cd %s
export PROJECT_ROOT=%s
exec %s -tool automation -args %s
`, shellQuote(cfg.ProjectRoot), shellQuote(cfg.ProjectRoot), shellQuote(cfg.BinaryPath), shellQuote(string(args)))

	artifacts := &automationScheduleArtifacts{
		ScriptContents: scriptContents,
	}

	if runtime.GOOS == "darwin" {
		plist, err := renderLaunchdPlist(cfg)
		if err != nil {
			return nil, err
		}
		artifacts.PlistContents = plist
	} else if runtime.GOOS == "linux" {
		service, timer := renderSystemdUnit(cfg)
		artifacts.ServiceContents = service
		artifacts.TimerContents = timer
	}

	return artifacts, nil
}

func renderLaunchdPlist(cfg *automationScheduleConfig) (string, error) {
	var out strings.Builder
	out.WriteString(`<?xml version="1.0" encoding="UTF-8"?>` + "\n")
	out.WriteString(`<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">` + "\n")
	out.WriteString(`<plist version="1.0">` + "\n")
	out.WriteString("  <dict>\n")
	writePlistKeyValue(&out, "Label", cfg.ScheduleLabel)
	writePlistArray(&out, "ProgramArguments", []string{cfg.ScriptPath})
	writePlistKeyValue(&out, "WorkingDirectory", cfg.ProjectRoot)
	writePlistKeyBool(&out, "RunAtLoad", cfg.RunAtLoad)
	writePlistKeyBool(&out, "Disabled", !cfg.Enabled)
	writePlistKeyBool(&out, "KeepAlive", false)
	writePlistKeyInt(&out, "StartInterval", cfg.IntervalSeconds)
	writePlistKeyValue(&out, "StandardOutPath", filepath.Join(filepath.Dir(cfg.LogPath), cfg.ScheduleLabel+".out.log"))
	writePlistKeyValue(&out, "StandardErrorPath", filepath.Join(filepath.Dir(cfg.LogPath), cfg.ScheduleLabel+".err.log"))
	out.WriteString("  </dict>\n")
	out.WriteString("</plist>\n")

	return out.String(), nil
}

func renderSystemdUnit(cfg *automationScheduleConfig) (service, timer string) {
	service = fmt.Sprintf(`[Unit]
Description=exarp-go automation %s

[Service]
Type=oneshot
ExecStart=%s
`, cfg.TargetAction, systemdEscapePath(cfg.ScriptPath))

	timer = fmt.Sprintf(`[Unit]
Description=exarp-go automation %s timer

[Timer]
OnBootSec=2m
OnUnitActiveSec=%dsec
Persistent=true
Unit=%s.service

[Install]
WantedBy=timers.target
`, cfg.TargetAction, cfg.IntervalSeconds, cfg.ScheduleLabel)

	return service, timer
}

func writeAutomationScheduleArtifacts(cfg *automationScheduleConfig, artifacts *automationScheduleArtifacts) error {
	if err := os.MkdirAll(filepath.Dir(cfg.ScriptPath), 0o755); err != nil {
		return fmt.Errorf("failed to create schedule script dir: %w", err)
	}
	if err := os.WriteFile(cfg.ScriptPath, []byte(artifacts.ScriptContents), 0o755); err != nil {
		return fmt.Errorf("failed to write schedule script: %w", err)
	}

	switch runtime.GOOS {
	case "darwin":
		if err := os.MkdirAll(filepath.Dir(cfg.PlistPath), 0o755); err != nil {
			return fmt.Errorf("failed to create LaunchAgents dir: %w", err)
		}
		if err := os.WriteFile(cfg.PlistPath, []byte(artifacts.PlistContents), 0o644); err != nil {
			return fmt.Errorf("failed to write launchd plist: %w", err)
		}
	case "linux":
		if err := os.MkdirAll(filepath.Dir(cfg.ServicePath), 0o755); err != nil {
			return fmt.Errorf("failed to create systemd dir: %w", err)
		}
		if err := os.WriteFile(cfg.ServicePath, []byte(artifacts.ServiceContents), 0o644); err != nil {
			return fmt.Errorf("failed to write systemd service: %w", err)
		}
		if err := os.WriteFile(cfg.TimerPath, []byte(artifacts.TimerContents), 0o644); err != nil {
			return fmt.Errorf("failed to write systemd timer: %w", err)
		}
	default:
		return nil
	}

	return nil
}

func removeAutomationScheduleArtifacts(cfg *automationScheduleConfig) (bool, error) {
	removed := false
	for _, path := range []string{cfg.ScriptPath, cfg.PlistPath, cfg.ServicePath, cfg.TimerPath} {
		if path == "" {
			continue
		}
		if err := os.Remove(path); err == nil {
			removed = true
		} else if !os.IsNotExist(err) {
			return removed, fmt.Errorf("failed to remove %s: %w", path, err)
		}
	}
	return removed, nil
}

func loadLaunchdJob(cfg *automationScheduleConfig) (bool, error) {
	domain := fmt.Sprintf("gui/%d", currentUID())
	if _, err := exec.LookPath("launchctl"); err != nil {
		return false, err
	}

	if err := exec.Command("launchctl", "bootstrap", domain, cfg.PlistPath).Run(); err == nil {
		return true, nil
	}

	if err := exec.Command("launchctl", "load", "-w", cfg.PlistPath).Run(); err != nil {
		return false, fmt.Errorf("failed to load launchd job: %w", err)
	}

	return true, nil
}

func unloadLaunchdJob(cfg *automationScheduleConfig) (bool, error) {
	domain := fmt.Sprintf("gui/%d", currentUID())
	if _, err := exec.LookPath("launchctl"); err != nil {
		return false, err
	}

	if err := exec.Command("launchctl", "bootout", domain, cfg.PlistPath).Run(); err == nil {
		return true, nil
	}

	if err := exec.Command("launchctl", "unload", "-w", cfg.PlistPath).Run(); err != nil {
		return false, fmt.Errorf("failed to unload launchd job: %w", err)
	}

	return true, nil
}

func loadSystemdTimer(cfg *automationScheduleConfig) (bool, error) {
	if _, err := exec.LookPath("systemctl"); err != nil {
		return false, err
	}

	if err := exec.Command("systemctl", "--user", "daemon-reload").Run(); err != nil {
		return false, fmt.Errorf("systemd daemon-reload failed: %w", err)
	}
	if err := exec.Command("systemctl", "--user", "enable", "--now", cfg.ScheduleLabel+".timer").Run(); err != nil {
		return false, fmt.Errorf("failed to enable systemd timer: %w", err)
	}

	return true, nil
}

func unloadSystemdTimer(cfg *automationScheduleConfig) (bool, error) {
	if _, err := exec.LookPath("systemctl"); err != nil {
		return false, err
	}

	_ = exec.Command("systemctl", "--user", "disable", "--now", cfg.ScheduleLabel+".timer").Run()
	return true, nil
}

func beginAutomationRun(ctx context.Context, action, scheduleLabel string) (*automationRunGuard, map[string]interface{}, error) {
	if scheduleLabel == "" {
		scheduleLabel = action
	}
	scheduleLabel = sanitizeAutomationLabel(scheduleLabel)

	var guard *automationRunGuard
	var skipResult map[string]interface{}

	err := database.WithRetry(ctx, func() error {
		db, err := database.GetDBx()
		if err != nil {
			return err
		}

		queryCtx, cancel := context.WithTimeout(ctx, databaseTimeout())
		defer cancel()

		tx, err := db.BeginTxx(queryCtx, nil)
		if err != nil {
			return err
		}
		defer func() {
			_ = tx.Rollback()
		}()

		if err := ensureAutomationRunsTableTx(queryCtx, tx); err != nil {
			return err
		}

		active, err := getActiveAutomationRunTx(queryCtx, tx, scheduleLabel)
		if err != nil {
			return err
		}
		if active != nil {
			if processRunning(active.PID) {
				_ = tx.Commit()
				skipResult = map[string]interface{}{
					"status":         "skipped",
					"reason":         "previous run still active",
					"schedule_label": scheduleLabel,
					"active_run":     active,
				}
				return nil
			}

			if _, err := tx.ExecContext(queryCtx, `
				UPDATE automation_runs
				SET status = 'stale',
					ended_at = ?,
					updated_at = ?,
					error_text = ?
				WHERE id = ?`,
				time.Now().Unix(),
				time.Now().Unix(),
				sql.NullString{String: "stale automation run replaced", Valid: true},
				active.ID,
			); err != nil {
				return err
			}
		}

		now := time.Now().Unix()
		host, _ := os.Hostname()
		res, err := tx.ExecContext(queryCtx, `
			INSERT INTO automation_runs (
				schedule_label, action, pid, host, status, started_at, created_at, updated_at
			) VALUES (?, ?, ?, ?, 'running', ?, ?, ?)
		`, scheduleLabel, action, os.Getpid(), host, now, now, now)
		if err != nil {
			if isUniqueConstraintError(err) {
				existing, getErr := getActiveAutomationRunTx(queryCtx, tx, scheduleLabel)
				if getErr != nil {
					return getErr
				}
				if existing != nil && processRunning(existing.PID) {
					_ = tx.Commit()
					skipResult = map[string]interface{}{
						"status":         "skipped",
						"reason":         "previous run still active",
						"schedule_label": scheduleLabel,
						"active_run":     existing,
					}
					return nil
				}
			}
			return err
		}

		runID, _ := res.LastInsertId()
		if err := tx.Commit(); err != nil {
			return err
		}

		guard = &automationRunGuard{
			RunID:         runID,
			ScheduleLabel: scheduleLabel,
			Action:        action,
			PID:           os.Getpid(),
			StartedAt:     time.Now(),
		}

		return nil
	})
	if err != nil {
		return nil, nil, err
	}

	if skipResult != nil {
		return nil, skipResult, nil
	}

	return guard, nil, nil
}

func finishAutomationRun(ctx context.Context, guard *automationRunGuard, status string, errText string) {
	if guard == nil || guard.RunID <= 0 {
		return
	}

	db, err := database.GetDBx()
	if err != nil {
		return
	}

	queryCtx, cancel := context.WithTimeout(ctx, databaseTimeout())
	defer cancel()

	_, _ = db.ExecContext(queryCtx, `
		UPDATE automation_runs
		SET status = ?, ended_at = ?, updated_at = ?, error_text = ?
		WHERE id = ?
	`, status, time.Now().Unix(), time.Now().Unix(), nullableString(errText), guard.RunID)
}

func ensureAutomationRunsTableTx(ctx context.Context, tx sqlxTx) error {
	_, err := tx.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS automation_runs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			schedule_label TEXT NOT NULL,
			action TEXT NOT NULL,
			pid INTEGER NOT NULL,
			host TEXT,
			status TEXT NOT NULL,
			started_at INTEGER NOT NULL,
			ended_at INTEGER,
			error_text TEXT,
			created_at INTEGER NOT NULL,
			updated_at INTEGER NOT NULL
		)
	`)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `CREATE INDEX IF NOT EXISTS idx_automation_runs_pid ON automation_runs(pid)`)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `CREATE INDEX IF NOT EXISTS idx_automation_runs_action_status ON automation_runs(action, status)`)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `CREATE UNIQUE INDEX IF NOT EXISTS idx_automation_runs_active_label ON automation_runs(schedule_label) WHERE status = 'running'`)
	if err != nil {
		return err
	}
	return nil
}

type sqlxTx interface {
	ExecContext(context.Context, string, ...any) (sql.Result, error)
	GetContext(context.Context, any, string, ...any) error
	SelectContext(context.Context, any, string, ...any) error
	Commit() error
	Rollback() error
}

func getActiveAutomationRunTx(ctx context.Context, tx sqlxTx, scheduleLabel string) (*automationRunRow, error) {
	var row automationRunRow
	err := tx.GetContext(ctx, &row, `
		SELECT id, schedule_label, action, pid, status, started_at, ended_at, error_text
		FROM automation_runs
		WHERE schedule_label = ? AND status = 'running'
		ORDER BY started_at DESC
		LIMIT 1
	`, scheduleLabel)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &row, nil
}

func processRunning(pid int) bool {
	if pid <= 0 {
		return false
	}

	proc, err := os.FindProcess(pid)
	if err != nil {
		return false
	}

	if runtime.GOOS == "windows" {
		return true
	}

	return proc.Signal(syscall.Signal(0)) == nil
}

func databaseTimeout() time.Duration { return 30 * time.Second }

func nullableString(s string) sql.NullString {
	if strings.TrimSpace(s) == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}

func scheduleResponseToMap(action string, result map[string]interface{}) map[string]interface{} {
	resp := &proto.AutomationResponse{Action: action}
	responseData := map[string]interface{}{
		"status":  "success",
		"results": result,
	}
	if payload, err := json.Marshal(responseData); err == nil {
		resp.ResultJson = string(payload)
	}
	return AutomationResponseToMap(resp)
}

func isAutomationTargetAction(action string) bool {
	switch action {
	case "daily", "nightly", "sprint", "discover":
		return true
	case "execution_cockpit":
		return true
	default:
		return false
	}
}

func defaultAutomationScheduleLabel(action string) string {
	return defaultAutomationLabelPrefix + "." + sanitizeAutomationLabel(action)
}

func sanitizeAutomationLabel(label string) string {
	label = strings.TrimSpace(label)
	if label == "" {
		return "automation"
	}

	var b strings.Builder
	lastDash := false
	for _, r := range label {
		switch {
		case r >= 'a' && r <= 'z':
			b.WriteRune(r)
			lastDash = false
		case r >= 'A' && r <= 'Z':
			b.WriteRune(r + ('a' - 'A'))
			lastDash = false
		case r >= '0' && r <= '9':
			b.WriteRune(r)
			lastDash = false
		case r == '.' || r == '-' || r == '_':
			if !lastDash {
				b.WriteRune('-')
				lastDash = true
			}
		default:
			if !lastDash {
				b.WriteRune('-')
				lastDash = true
			}
		}
	}

	out := strings.Trim(b.String(), "-")
	if out == "" {
		return "automation"
	}
	return out
}

func shellQuote(s string) string {
	return strconv.Quote(s)
}

func systemdEscapePath(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, " ", `\x20`)
	return s
}

func currentUID() int {
	current, err := user.Current()
	if err != nil {
		return 0
	}
	uid, err := strconv.Atoi(current.Uid)
	if err != nil {
		return 0
	}
	return uid
}

func writePlistKeyValue(out *strings.Builder, key, value string) {
	out.WriteString("    <key>")
	out.WriteString(xmlEscape(key))
	out.WriteString("</key>\n")
	out.WriteString("    <string>")
	out.WriteString(xmlEscape(value))
	out.WriteString("</string>\n")
}

func writePlistKeyBool(out *strings.Builder, key string, value bool) {
	out.WriteString("    <key>")
	out.WriteString(xmlEscape(key))
	out.WriteString("</key>\n")
	if value {
		out.WriteString("    <true/>\n")
		return
	}
	out.WriteString("    <false/>\n")
}

func writePlistKeyInt(out *strings.Builder, key string, value int) {
	out.WriteString("    <key>")
	out.WriteString(xmlEscape(key))
	out.WriteString("</key>\n")
	out.WriteString("    <integer>")
	out.WriteString(strconv.Itoa(value))
	out.WriteString("</integer>\n")
}

func writePlistArray(out *strings.Builder, key string, values []string) {
	out.WriteString("    <key>")
	out.WriteString(xmlEscape(key))
	out.WriteString("</key>\n")
	out.WriteString("    <array>\n")
	for _, v := range values {
		out.WriteString("      <string>")
		out.WriteString(xmlEscape(v))
		out.WriteString("</string>\n")
	}
	out.WriteString("    </array>\n")
}

func xmlEscape(s string) string {
	var b strings.Builder
	_ = xml.EscapeText(&b, []byte(s))
	return b.String()
}
