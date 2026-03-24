package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"
)

type TaskExecutionRun struct {
	RunID        string    `json:"run_id"`
	TaskID       string    `json:"task_id"`
	AgentID      string    `json:"agent_id,omitempty"`
	Host         string    `json:"host,omitempty"`
	Status       string    `json:"status"`
	Summary      string    `json:"summary,omitempty"`
	FilesTouched []string  `json:"files_touched,omitempty"`
	CommandsRun  []string  `json:"commands_run,omitempty"`
	Notes        string    `json:"notes,omitempty"`
	StartedAt    time.Time `json:"started_at"`
	EndedAt      time.Time `json:"ended_at,omitempty"`
}

type TaskVerification struct {
	VerificationID string    `json:"verification_id"`
	TaskID         string    `json:"task_id"`
	RunID          string    `json:"run_id,omitempty"`
	Kind           string    `json:"kind"`
	Command        string    `json:"command,omitempty"`
	Result         string    `json:"result"`
	Details        string    `json:"details,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
}

type TaskProgressEntry struct {
	ProgressID    string    `json:"progress_id"`
	TaskID        string    `json:"task_id"`
	RunID         string    `json:"run_id,omitempty"`
	Summary       string    `json:"summary"`
	Files         []string  `json:"files,omitempty"`
	RemainingWork string    `json:"remaining_work,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
}

type taskExecutionRunRow struct {
	RunID        string        `db:"run_id"`
	TaskID       string        `db:"task_id"`
	AgentID      string        `db:"agent_id"`
	Host         string        `db:"host"`
	Status       string        `db:"status"`
	Summary      string        `db:"summary"`
	FilesTouched string        `db:"files_touched"`
	CommandsRun  string        `db:"commands_run"`
	Notes        string        `db:"notes"`
	StartedAt    int64         `db:"started_at"`
	EndedAt      sql.NullInt64 `db:"ended_at"`
}

func GenerateExecutionRunID() string {
	return fmt.Sprintf("R-%d", time.Now().UnixNano())
}

func GenerateVerificationID() string {
	return fmt.Sprintf("V-%d", time.Now().UnixNano())
}

func GenerateProgressID() string {
	return fmt.Sprintf("P-%d", time.Now().UnixNano())
}

func StartTaskExecutionRun(ctx context.Context, run *TaskExecutionRun) error {
	if run == nil {
		return fmt.Errorf("run is required")
	}
	if strings.TrimSpace(run.TaskID) == "" {
		return fmt.Errorf("task_id is required")
	}
	if strings.TrimSpace(run.RunID) == "" {
		run.RunID = GenerateExecutionRunID()
	}
	if strings.TrimSpace(run.Status) == "" {
		run.Status = "running"
	}
	if run.StartedAt.IsZero() {
		run.StartedAt = time.Now()
	}
	if run.Host == "" {
		if host, err := os.Hostname(); err == nil {
			run.Host = host
		}
	}

	filesJSON, err := marshalStringList(run.FilesTouched)
	if err != nil {
		return err
	}
	commandsJSON, err := marshalStringList(run.CommandsRun)
	if err != nil {
		return err
	}

	return execWithRetry(ctx, func(execCtx context.Context) error {
		db, err := GetDBx()
		if err != nil {
			return fmt.Errorf("failed to get database: %w", err)
		}
		_, err = db.ExecContext(execCtx, `
            INSERT INTO task_execution_runs (
                run_id, task_id, agent_id, host, status, summary, files_touched, commands_run, notes, started_at, created_at, updated_at
            ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, strftime('%s', 'now'), strftime('%s', 'now'))
        `, run.RunID, run.TaskID, run.AgentID, run.Host, run.Status, run.Summary, filesJSON, commandsJSON, run.Notes, run.StartedAt.Unix())
		if err != nil {
			return fmt.Errorf("failed to insert execution run: %w", err)
		}
		return nil
	})
}

func EndTaskExecutionRun(ctx context.Context, runID, status, summary string, filesTouched, commandsRun []string, notes string) error {
	if strings.TrimSpace(runID) == "" {
		return fmt.Errorf("run_id is required")
	}
	if strings.TrimSpace(status) == "" {
		status = "completed"
	}
	filesJSON, err := marshalStringList(filesTouched)
	if err != nil {
		return err
	}
	commandsJSON, err := marshalStringList(commandsRun)
	if err != nil {
		return err
	}

	return execWithRetry(ctx, func(execCtx context.Context) error {
		db, err := GetDBx()
		if err != nil {
			return fmt.Errorf("failed to get database: %w", err)
		}
		result, err := db.ExecContext(execCtx, `
            UPDATE task_execution_runs
            SET status = ?, summary = CASE WHEN ? <> '' THEN ? ELSE summary END,
                files_touched = CASE WHEN ? <> '[]' THEN ? ELSE files_touched END,
                commands_run = CASE WHEN ? <> '[]' THEN ? ELSE commands_run END,
                notes = CASE WHEN ? <> '' THEN ? ELSE notes END,
                ended_at = ?, updated_at = strftime('%s', 'now')
            WHERE run_id = ?
        `, status, summary, summary, filesJSON, filesJSON, commandsJSON, commandsJSON, notes, notes, time.Now().Unix(), runID)
		if err != nil {
			return fmt.Errorf("failed to update execution run: %w", err)
		}
		rows, err := result.RowsAffected()
		if err != nil {
			return fmt.Errorf("failed to check updated execution run: %w", err)
		}
		if rows == 0 {
			return fmt.Errorf("execution run %s not found", runID)
		}
		return nil
	})
}

func GetTaskExecutionRun(ctx context.Context, runID string) (*TaskExecutionRun, error) {
	if strings.TrimSpace(runID) == "" {
		return nil, fmt.Errorf("run_id is required")
	}

	var row taskExecutionRunRow
	err := queryWithRetry(ctx, func(queryCtx context.Context) error {
		db, err := GetDBx()
		if err != nil {
			return fmt.Errorf("failed to get database: %w", err)
		}
		return db.GetContext(queryCtx, &row, `
            SELECT run_id, task_id, agent_id, host, status, summary, files_touched, commands_run, notes, started_at, ended_at
            FROM task_execution_runs
            WHERE run_id = ?
        `, runID)
	})
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("execution run %s not found", runID)
		}
		return nil, err
	}
	return executionRunFromRow(row), nil
}

func ListTaskExecutionRuns(ctx context.Context, taskID, status string, limit int) ([]TaskExecutionRun, error) {
	var rows []taskExecutionRunRow

	err := queryWithRetry(ctx, func(queryCtx context.Context) error {
		db, err := GetDBx()
		if err != nil {
			return fmt.Errorf("failed to get database: %w", err)
		}
		query := `
            SELECT run_id, task_id, agent_id, host, status, summary, files_touched, commands_run, notes, started_at, ended_at
            FROM task_execution_runs
            WHERE (? = '' OR task_id = ?)
              AND (? = '' OR status = ?)
            ORDER BY started_at DESC
        `
		args := []interface{}{taskID, taskID, status, status}
		if limit > 0 {
			query += " LIMIT ?"
			args = append(args, limit)
		}
		return db.SelectContext(queryCtx, &rows, query, args...)
	})
	if err != nil {
		return nil, err
	}

	runs := make([]TaskExecutionRun, 0, len(rows))
	for _, row := range rows {
		runs = append(runs, *executionRunFromRow(row))
	}
	return runs, nil
}

func AddTaskVerification(ctx context.Context, verification *TaskVerification) error {
	if verification == nil {
		return fmt.Errorf("verification is required")
	}
	if strings.TrimSpace(verification.TaskID) == "" {
		return fmt.Errorf("task_id is required")
	}
	if strings.TrimSpace(verification.Kind) == "" {
		return fmt.Errorf("kind is required")
	}
	if strings.TrimSpace(verification.Result) == "" {
		return fmt.Errorf("result is required")
	}
	if verification.VerificationID == "" {
		verification.VerificationID = GenerateVerificationID()
	}
	if verification.CreatedAt.IsZero() {
		verification.CreatedAt = time.Now()
	}

	return execWithRetry(ctx, func(execCtx context.Context) error {
		db, err := GetDBx()
		if err != nil {
			return fmt.Errorf("failed to get database: %w", err)
		}
		_, err = db.ExecContext(execCtx, `
            INSERT INTO task_verifications (verification_id, task_id, run_id, kind, command, result, details, created_at)
            VALUES (?, ?, ?, ?, ?, ?, ?, ?)
        `, verification.VerificationID, verification.TaskID, verification.RunID, verification.Kind, verification.Command, verification.Result, verification.Details, verification.CreatedAt.Unix())
		if err != nil {
			return fmt.Errorf("failed to insert verification: %w", err)
		}
		return nil
	})
}

func ListTaskVerifications(ctx context.Context, taskID, runID string, limit int) ([]TaskVerification, error) {
	var rows []struct {
		VerificationID string `db:"verification_id"`
		TaskID         string `db:"task_id"`
		RunID          string `db:"run_id"`
		Kind           string `db:"kind"`
		Command        string `db:"command"`
		Result         string `db:"result"`
		Details        string `db:"details"`
		CreatedAt      int64  `db:"created_at"`
	}
	err := queryWithRetry(ctx, func(queryCtx context.Context) error {
		db, err := GetDBx()
		if err != nil {
			return fmt.Errorf("failed to get database: %w", err)
		}
		query := `
            SELECT verification_id, task_id, run_id, kind, command, result, details, created_at
            FROM task_verifications
            WHERE (? = '' OR task_id = ?)
              AND (? = '' OR run_id = ?)
            ORDER BY created_at DESC
        `
		args := []interface{}{taskID, taskID, runID, runID}
		if limit > 0 {
			query += " LIMIT ?"
			args = append(args, limit)
		}
		return db.SelectContext(queryCtx, &rows, query, args...)
	})
	if err != nil {
		return nil, err
	}
	out := make([]TaskVerification, 0, len(rows))
	for _, row := range rows {
		out = append(out, TaskVerification{
			VerificationID: row.VerificationID,
			TaskID:         row.TaskID,
			RunID:          row.RunID,
			Kind:           row.Kind,
			Command:        row.Command,
			Result:         row.Result,
			Details:        row.Details,
			CreatedAt:      time.Unix(row.CreatedAt, 0),
		})
	}
	return out, nil
}

func AddTaskProgressEntry(ctx context.Context, entry *TaskProgressEntry) error {
	if entry == nil {
		return fmt.Errorf("progress entry is required")
	}
	if strings.TrimSpace(entry.TaskID) == "" {
		return fmt.Errorf("task_id is required")
	}
	if strings.TrimSpace(entry.Summary) == "" {
		return fmt.Errorf("summary is required")
	}
	if entry.ProgressID == "" {
		entry.ProgressID = GenerateProgressID()
	}
	if entry.CreatedAt.IsZero() {
		entry.CreatedAt = time.Now()
	}
	filesJSON, err := marshalStringList(entry.Files)
	if err != nil {
		return err
	}

	return execWithRetry(ctx, func(execCtx context.Context) error {
		db, err := GetDBx()
		if err != nil {
			return fmt.Errorf("failed to get database: %w", err)
		}
		_, err = db.ExecContext(execCtx, `
            INSERT INTO task_progress_entries (progress_id, task_id, run_id, summary, files, remaining_work, created_at)
            VALUES (?, ?, ?, ?, ?, ?, ?)
        `, entry.ProgressID, entry.TaskID, entry.RunID, entry.Summary, filesJSON, entry.RemainingWork, entry.CreatedAt.Unix())
		if err != nil {
			return fmt.Errorf("failed to insert progress entry: %w", err)
		}
		return nil
	})
}

func ListTaskProgressEntries(ctx context.Context, taskID, runID string, limit int) ([]TaskProgressEntry, error) {
	var rows []struct {
		ProgressID    string `db:"progress_id"`
		TaskID        string `db:"task_id"`
		RunID         string `db:"run_id"`
		Summary       string `db:"summary"`
		Files         string `db:"files"`
		RemainingWork string `db:"remaining_work"`
		CreatedAt     int64  `db:"created_at"`
	}
	err := queryWithRetry(ctx, func(queryCtx context.Context) error {
		db, err := GetDBx()
		if err != nil {
			return fmt.Errorf("failed to get database: %w", err)
		}
		query := `
            SELECT progress_id, task_id, run_id, summary, files, remaining_work, created_at
            FROM task_progress_entries
            WHERE (? = '' OR task_id = ?)
              AND (? = '' OR run_id = ?)
            ORDER BY created_at DESC
        `
		args := []interface{}{taskID, taskID, runID, runID}
		if limit > 0 {
			query += " LIMIT ?"
			args = append(args, limit)
		}
		return db.SelectContext(queryCtx, &rows, query, args...)
	})
	if err != nil {
		return nil, err
	}
	out := make([]TaskProgressEntry, 0, len(rows))
	for _, row := range rows {
		out = append(out, TaskProgressEntry{
			ProgressID:    row.ProgressID,
			TaskID:        row.TaskID,
			RunID:         row.RunID,
			Summary:       row.Summary,
			Files:         unmarshalStringList(row.Files),
			RemainingWork: row.RemainingWork,
			CreatedAt:     time.Unix(row.CreatedAt, 0),
		})
	}
	return out, nil
}

func execWithRetry(ctx context.Context, fn func(execCtx context.Context) error) error {
	ctx = ensureContext(ctx)
	execCtx, cancel := withTransactionTimeout(ctx)
	defer cancel()
	return retryWithBackoff(ctx, func() error {
		return fn(execCtx)
	})
}

func queryWithRetry(ctx context.Context, fn func(queryCtx context.Context) error) error {
	ctx = ensureContext(ctx)
	queryCtx, cancel := withQueryTimeout(ctx)
	defer cancel()
	return retryWithBackoff(ctx, func() error {
		return fn(queryCtx)
	})
}

func marshalStringList(items []string) (string, error) {
	if len(items) == 0 {
		return "[]", nil
	}
	b, err := json.Marshal(items)
	if err != nil {
		return "", fmt.Errorf("failed to marshal string list: %w", err)
	}
	return string(b), nil
}

func unmarshalStringList(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	var out []string
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return nil
	}
	return out
}

func executionRunFromRow(r taskExecutionRunRow) *TaskExecutionRun {
	run := &TaskExecutionRun{
		RunID:        r.RunID,
		TaskID:       r.TaskID,
		AgentID:      r.AgentID,
		Host:         r.Host,
		Status:       r.Status,
		Summary:      r.Summary,
		FilesTouched: unmarshalStringList(r.FilesTouched),
		CommandsRun:  unmarshalStringList(r.CommandsRun),
		Notes:        r.Notes,
		StartedAt:    time.Unix(r.StartedAt, 0),
	}
	if ended := r.EndedAt; ended.Valid {
		run.EndedAt = time.Unix(ended.Int64, 0)
	}
	return run
}
