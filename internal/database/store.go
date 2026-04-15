package database

import (
	"context"
)

// TaskStore abstracts task persistence for testing and DB/file fallback.
// Implementations may use the database, JSON file, or an in-memory mock.
type TaskStore interface {
	GetTask(ctx context.Context, id string) (*Todo2Task, error)
	UpdateTask(ctx context.Context, task *Todo2Task) error
	ListTasks(ctx context.Context, filters *TaskFilters) ([]*Todo2Task, error)
	CreateTask(ctx context.Context, task *Todo2Task) error
	DeleteTask(ctx context.Context, id string) error
}

// dbStore delegates to the package-level database functions.
// Use when the database has been initialized (e.g. Init(projectRoot) was called).
type dbStore struct{}

// DefaultDBStore is a TaskStore that uses the global database connection.
// Returns errors if the database is not initialized.
var DefaultDBStore TaskStore = &dbStore{}

func (dbStore) GetTask(ctx context.Context, id string) (*Todo2Task, error) {
	return GetTask(ctx, id)
}

func (dbStore) UpdateTask(ctx context.Context, task *Todo2Task) error {
	return UpdateTask(ctx, task)
}

func (dbStore) ListTasks(ctx context.Context, filters *TaskFilters) ([]*Todo2Task, error) {
	return ListTasks(ctx, filters)
}

func (dbStore) CreateTask(ctx context.Context, task *Todo2Task) error {
	return CreateTask(ctx, task)
}

func (dbStore) DeleteTask(ctx context.Context, id string) error {
	return DeleteTask(ctx, id)
}
