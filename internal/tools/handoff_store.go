// handoff_store.go — Typed session handoff persistence: single decode into HandoffStore,
// on-disk format gzip+gob after magic (efficient vs indented JSON). Legacy handoffs.json
// is read once and migrated on next write.
package tools

import (
	"bytes"
	"compress/gzip"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/davidl71/exarp-go/internal/cache"
	"github.com/spf13/cast"
)

const (
	handoffsStoreFile      = "handoffs.store"
	handoffsLegacyJSONFile = "handoffs.json"
)

// handoffStoreMagic identifies the binary store format (version 1).
var handoffStoreMagic = []byte("EXARPHF\x01")

// HandoffStore is the root document for persisted handoffs (JSON or gob).
type HandoffStore struct {
	Handoffs []HandoffEntry `json:"handoffs"`
}

// HandoffEntry is one session handoff note (matches former handoffs.json entry shape).
type HandoffEntry struct {
	ID                           string                  `json:"id"`
	Timestamp                    string                  `json:"timestamp"`
	Host                         string                  `json:"host"`
	Summary                      string                  `json:"summary"`
	Blockers                     []string                `json:"blockers,omitempty"`
	NextSteps                    []string                `json:"next_steps,omitempty"`
	GitStatus                    *GitStatusHandoff       `json:"git_status,omitempty"`
	TasksInProgress              []TaskInProgressHandoff `json:"tasks_in_progress,omitempty"`
	TaskJournal                  []TaskJournalEntry      `json:"task_journal,omitempty"`
	PointInTimeSnapshot          string                  `json:"point_in_time_snapshot,omitempty"`
	PointInTimeSnapshotFormat    string                  `json:"point_in_time_snapshot_format,omitempty"`
	PointInTimeSnapshotTaskCount int                     `json:"point_in_time_snapshot_task_count,omitempty"`
	LedgerWriteWarning           string                  `json:"ledger_write_warning,omitempty"`
	ContinuityLedgerPath         string                  `json:"continuity_ledger_path,omitempty"`
	Status                       string                  `json:"status,omitempty"`
}

// GitStatusHandoff mirrors getGitStatus map shape.
type GitStatusHandoff struct {
	Branch           string   `json:"branch,omitempty"`
	UncommittedFiles int      `json:"uncommitted_files,omitempty"`
	ChangedFiles     []string `json:"changed_files,omitempty"`
}

// TaskInProgressHandoff is a minimal task summary embedded in a handoff.
type TaskInProgressHandoff struct {
	ID      string `json:"id"`
	Content string `json:"content"`
	Status  string `json:"status"`
}

// TaskJournalEntry records task activity during a session.
type TaskJournalEntry struct {
	ID      string `json:"id,omitempty"`
	Action  string `json:"action,omitempty"`
	Summary string `json:"summary,omitempty"`
}

func handoffsStorePath(projectRoot string) string {
	return filepath.Join(projectRoot, ".todo2", handoffsStoreFile)
}

func handoffsLegacyJSONPath(projectRoot string) string {
	return filepath.Join(projectRoot, ".todo2", handoffsLegacyJSONFile)
}

// handoffsPersistPath returns the path to an existing handoff file (store preferred over legacy JSON).
func handoffsPersistPath(projectRoot string) string {
	p := handoffsStorePath(projectRoot)
	if _, err := os.Stat(p); err == nil {
		return p
	}
	return handoffsLegacyJSONPath(projectRoot)
}

// handoffsAnyFileExists reports whether either the binary store or legacy JSON exists.
func handoffsAnyFileExists(projectRoot string) bool {
	_, err1 := os.Stat(handoffsStorePath(projectRoot))
	_, err2 := os.Stat(handoffsLegacyJSONPath(projectRoot))
	return err1 == nil || err2 == nil
}

// loadHandoffStoreFromBytes decodes store bytes (binary magic+gzip+gob or legacy JSON) in one pass.
func loadHandoffStoreFromBytes(_ string, data []byte) (HandoffStore, error) {
	if len(data) >= len(handoffStoreMagic) && bytes.HasPrefix(data, handoffStoreMagic) {
		return decodeHandoffStoreBinary(data)
	}
	var s HandoffStore
	if err := json.Unmarshal(data, &s); err != nil {
		return HandoffStore{}, err
	}
	return s, nil
}

// loadHandoffStore loads the handoff history with a single structured decode per file:
// binary (magic + gzip + gob) or legacy JSON into HandoffStore.
func loadHandoffStore(projectRoot string) (HandoffStore, error) {
	fc := cache.GetGlobalFileCache()
	storePath := handoffsStorePath(projectRoot)
	if data, _, err := fc.ReadFile(storePath); err == nil {
		s, err := loadHandoffStoreFromBytes(storePath, data)
		if err != nil {
			return HandoffStore{}, fmt.Errorf("handoffs.store: %w", err)
		}
		return s, nil
	}

	jsonPath := handoffsLegacyJSONPath(projectRoot)
	data, _, err := fc.ReadFile(jsonPath)
	if err != nil {
		if os.IsNotExist(err) {
			return HandoffStore{}, nil
		}
		return HandoffStore{}, err
	}
	s, err := loadHandoffStoreFromBytes(jsonPath, data)
	if err != nil {
		return HandoffStore{}, fmt.Errorf("parse legacy handoffs.json: %w", err)
	}
	return s, nil
}

func decodeHandoffStoreBinary(data []byte) (HandoffStore, error) {
	if len(data) < len(handoffStoreMagic) || !bytes.HasPrefix(data, handoffStoreMagic) {
		return HandoffStore{}, fmt.Errorf("invalid handoff store magic or truncated file")
	}
	r, err := gzip.NewReader(bytes.NewReader(data[len(handoffStoreMagic):]))
	if err != nil {
		return HandoffStore{}, fmt.Errorf("handoff store gzip: %w", err)
	}
	defer r.Close()

	dec := gob.NewDecoder(r)
	var s HandoffStore
	if err := dec.Decode(&s); err != nil {
		return HandoffStore{}, fmt.Errorf("handoff store gob decode: %w", err)
	}
	return s, nil
}

// saveHandoffStore writes HandoffStore as magic + gzip + gob and removes legacy handoffs.json if present.
func saveHandoffStore(projectRoot string, store HandoffStore) error {
	dir := filepath.Join(projectRoot, ".todo2")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	var payload bytes.Buffer
	gw := gzip.NewWriter(&payload)
	enc := gob.NewEncoder(gw)
	if err := enc.Encode(&store); err != nil {
		_ = gw.Close()
		return fmt.Errorf("handoff store gob encode: %w", err)
	}
	if err := gw.Close(); err != nil {
		return fmt.Errorf("handoff store gzip close: %w", err)
	}

	storePath := handoffsStorePath(projectRoot)
	tmp := storePath + ".tmp"
	f, err := os.OpenFile(tmp, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}
	closeRemove := func() {
		_ = f.Close()
		_ = os.Remove(tmp)
	}
	if _, err := f.Write(handoffStoreMagic); err != nil {
		closeRemove()
		return err
	}
	if _, err := f.Write(payload.Bytes()); err != nil {
		closeRemove()
		return err
	}
	if err := f.Close(); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	if err := os.Rename(tmp, storePath); err != nil {
		_ = os.Remove(tmp)
		return err
	}

	legacy := handoffsLegacyJSONPath(projectRoot)
	_ = os.Remove(legacy)
	return nil
}

// handoffEntryToMap projects HandoffEntry into map[string]interface{} for MCP responses.
// It mirrors encoding/json omitempty for each field and avoids a JSON marshal+unmarshal
// round-trip (no extra full-buffer copy of the handoff payload).
func handoffEntryToMap(e HandoffEntry) (map[string]interface{}, error) {
	m := make(map[string]interface{}, 24)
	if e.ID != "" {
		m["id"] = e.ID
	}
	if e.Timestamp != "" {
		m["timestamp"] = e.Timestamp
	}
	if e.Host != "" {
		m["host"] = e.Host
	}
	if e.Summary != "" {
		m["summary"] = e.Summary
	}
	if len(e.Blockers) > 0 {
		blockers := make([]interface{}, len(e.Blockers))
		for i, s := range e.Blockers {
			blockers[i] = s
		}
		m["blockers"] = blockers
	}
	if len(e.NextSteps) > 0 {
		steps := make([]interface{}, len(e.NextSteps))
		for i, s := range e.NextSteps {
			steps[i] = s
		}
		m["next_steps"] = steps
	}
	if e.GitStatus != nil {
		gs := make(map[string]interface{}, 4)
		if e.GitStatus.Branch != "" {
			gs["branch"] = e.GitStatus.Branch
		}
		if e.GitStatus.UncommittedFiles != 0 {
			gs["uncommitted_files"] = e.GitStatus.UncommittedFiles
		}
		if len(e.GitStatus.ChangedFiles) > 0 {
			cf := make([]interface{}, len(e.GitStatus.ChangedFiles))
			for i, s := range e.GitStatus.ChangedFiles {
				cf[i] = s
			}
			gs["changed_files"] = cf
		}
		m["git_status"] = gs
	}
	if len(e.TasksInProgress) > 0 {
		arr := make([]interface{}, len(e.TasksInProgress))
		for i, t := range e.TasksInProgress {
			tm := make(map[string]interface{}, 3)
			if t.ID != "" {
				tm["id"] = t.ID
			}
			if t.Content != "" {
				tm["content"] = t.Content
			}
			if t.Status != "" {
				tm["status"] = t.Status
			}
			arr[i] = tm
		}
		m["tasks_in_progress"] = arr
	}
	if len(e.TaskJournal) > 0 {
		arr := make([]interface{}, len(e.TaskJournal))
		for i, j := range e.TaskJournal {
			jm := make(map[string]interface{}, 3)
			if j.ID != "" {
				jm["id"] = j.ID
			}
			if j.Action != "" {
				jm["action"] = j.Action
			}
			if j.Summary != "" {
				jm["summary"] = j.Summary
			}
			arr[i] = jm
		}
		m["task_journal"] = arr
	}
	if e.PointInTimeSnapshot != "" {
		m["point_in_time_snapshot"] = e.PointInTimeSnapshot
	}
	if e.PointInTimeSnapshotFormat != "" {
		m["point_in_time_snapshot_format"] = e.PointInTimeSnapshotFormat
	}
	if e.PointInTimeSnapshotTaskCount != 0 {
		m["point_in_time_snapshot_task_count"] = e.PointInTimeSnapshotTaskCount
	}
	if e.LedgerWriteWarning != "" {
		m["ledger_write_warning"] = e.LedgerWriteWarning
	}
	if e.ContinuityLedgerPath != "" {
		m["continuity_ledger_path"] = e.ContinuityLedgerPath
	}
	if e.Status != "" {
		m["status"] = e.Status
	}
	return m, nil
}

func gitStatusFromMap(m map[string]interface{}) *GitStatusHandoff {
	if len(m) == 0 {
		return nil
	}
	g := &GitStatusHandoff{
		Branch:           cast.ToString(m["branch"]),
		UncommittedFiles: cast.ToInt(m["uncommitted_files"]),
	}
	if cf, ok := m["changed_files"].([]interface{}); ok {
		for _, v := range cf {
			if s, ok := v.(string); ok {
				g.ChangedFiles = append(g.ChangedFiles, s)
			}
		}
	}
	if cf, ok := m["changed_files"].([]string); ok {
		g.ChangedFiles = cf
	}
	return g
}

func tasksInProgressFromMaps(maps []map[string]interface{}) []TaskInProgressHandoff {
	out := make([]TaskInProgressHandoff, 0, len(maps))
	for _, m := range maps {
		out = append(out, TaskInProgressHandoff{
			ID:      GetString(m, "id"),
			Content: GetString(m, "content"),
			Status:  GetString(m, "status"),
		})
	}
	return out
}

func taskJournalFromMaps(j []map[string]interface{}) []TaskJournalEntry {
	out := make([]TaskJournalEntry, 0, len(j))
	for _, m := range j {
		out = append(out, TaskJournalEntry{
			ID:      GetString(m, "id"),
			Action:  GetString(m, "action"),
			Summary: GetString(m, "summary"),
		})
	}
	return out
}

func taskJournalToMaps(j []TaskJournalEntry) []map[string]interface{} {
	out := make([]map[string]interface{}, 0, len(j))
	for _, e := range j {
		m := map[string]interface{}{}
		if e.ID != "" {
			m["id"] = e.ID
		}
		if e.Action != "" {
			m["action"] = e.Action
		}
		if e.Summary != "" {
			m["summary"] = e.Summary
		}
		out = append(out, m)
	}
	return out
}

// latestHandoffSummary returns the summary line from the most recent handoff, or "".
func latestHandoffSummary(store HandoffStore) string {
	if len(store.Handoffs) == 0 {
		return ""
	}
	return store.Handoffs[len(store.Handoffs)-1].Summary
}
