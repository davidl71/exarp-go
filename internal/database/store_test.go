package database

import (
	"context"
	"sync"
	"testing"
)

// MockTaskStore is an in-memory TaskStore for tests.
type MockTaskStore struct {
	mu    sync.RWMutex
	tasks map[string]*Todo2Task
}

// NewMockTaskStore returns an empty mock store.
func NewMockTaskStore() *MockTaskStore {
	return &MockTaskStore{tasks: make(map[string]*Todo2Task)}
}

func (m *MockTaskStore) GetTask(ctx context.Context, id string) (*Todo2Task, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	t, ok := m.tasks[id]
	if !ok {
		return nil, nil
	}
	// Return a copy so callers don't mutate the store
	cp := *t

	return &cp, nil
}

func (m *MockTaskStore) UpdateTask(ctx context.Context, task *Todo2Task) error {
	if task == nil {
		return nil
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	m.tasks[task.ID] = task

	return nil
}

func (m *MockTaskStore) ListTasks(ctx context.Context, filters *TaskFilters) ([]*Todo2Task, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var out []*Todo2Task

	for _, t := range m.tasks {
		if filters != nil {
			if filters.Status != nil && t.Status != *filters.Status {
				continue
			}

			if filters.Priority != nil && t.Priority != *filters.Priority {
				continue
			}

			if filters.Tag != nil {
				found := false

				for _, tag := range t.Tags {
					if tag == *filters.Tag {
						found = true
						break
					}
				}

				if !found {
					continue
				}
			}
		}

		cp := *t
		out = append(out, &cp)
	}

	return out, nil
}

func (m *MockTaskStore) CreateTask(ctx context.Context, task *Todo2Task) error {
	if task == nil {
		return nil
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	m.tasks[task.ID] = task

	return nil
}

func (m *MockTaskStore) DeleteTask(ctx context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.tasks, id)

	return nil
}

// Ensure MockTaskStore and DefaultDBStore implement TaskStore at compile time.
var (
	_ TaskStore = (*MockTaskStore)(nil)
	_ TaskStore = DefaultDBStore
)

func TestDefaultDBStore_ImplementsTaskStore(t *testing.T) {
	tmpDir := t.TempDir()
	if err := Init(tmpDir); err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	defer Close()

	ctx := context.Background()
	task := &Todo2Task{
		ID:       "T-999001",
		Content:  "Store test",
		Status:   "Todo",
		Priority: "medium",
		Tags:     []string{"store"},
	}

	if err := DefaultDBStore.CreateTask(ctx, task); err != nil {
		t.Fatalf("CreateTask: %v", err)
	}

	got, err := DefaultDBStore.GetTask(ctx, "T-999001")
	if err != nil || got == nil {
		t.Fatalf("GetTask: err=%v got=%v", err, got)
	}

	if got.Content != "Store test" {
		t.Errorf("GetTask content = %q, want Store test", got.Content)
	}

	got.Status = "Done"
	if err := DefaultDBStore.UpdateTask(ctx, got); err != nil {
		t.Fatalf("UpdateTask: %v", err)
	}

	list, err := DefaultDBStore.ListTasks(ctx, nil)
	if err != nil || len(list) < 1 {
		t.Fatalf("ListTasks: err=%v len=%d", err, len(list))
	}

	if err := DefaultDBStore.DeleteTask(ctx, "T-999001"); err != nil {
		t.Fatalf("DeleteTask: %v", err)
	}

	_, err = DefaultDBStore.GetTask(ctx, "T-999001")
	if err == nil {
		t.Error("GetTask after delete should error or return nil")
	}
}

func TestMockTaskStore(t *testing.T) {
	ctx := context.Background()
	mock := NewMockTaskStore()

	task := &Todo2Task{
		ID:       "T-mock-1",
		Content:  "Mock task",
		Status:   "Todo",
		Priority: "high",
		Tags:     []string{"a", "b"},
	}
	if err := mock.CreateTask(ctx, task); err != nil {
		t.Fatalf("CreateTask: %v", err)
	}

	got, err := mock.GetTask(ctx, "T-mock-1")
	if err != nil || got == nil {
		t.Fatalf("GetTask: err=%v got=%v", err, got)
	}

	if got.Content != "Mock task" {
		t.Errorf("GetTask content = %q, want Mock task", got.Content)
	}

	status := "In Progress"

	list, err := mock.ListTasks(ctx, &TaskFilters{Status: &status})
	if err != nil || len(list) != 0 {
		t.Fatalf("ListTasks status=In Progress: err=%v len=%d (task is Todo)", err, len(list))
	}

	list, err = mock.ListTasks(ctx, nil)
	if err != nil || len(list) != 1 {
		t.Fatalf("ListTasks nil filters: err=%v len=%d", err, len(list))
	}

	got.Status = "Done"
	if err := mock.UpdateTask(ctx, got); err != nil {
		t.Fatalf("UpdateTask: %v", err)
	}

	got2, _ := mock.GetTask(ctx, "T-mock-1")
	if got2.Status != "Done" {
		t.Errorf("after UpdateTask Status = %q, want Done", got2.Status)
	}

	if err := mock.DeleteTask(ctx, "T-mock-1"); err != nil {
		t.Fatalf("DeleteTask: %v", err)
	}

	got3, _ := mock.GetTask(ctx, "T-mock-1")
	if got3 != nil {
		t.Errorf("GetTask after delete = %v, want nil", got3)
	}
}
