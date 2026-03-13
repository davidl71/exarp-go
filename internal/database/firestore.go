// firestore.go — Firestore TaskStore implementation (optional, requires cloud.google.com/go/firestore).
//go:build with_firestore
// +build with_firestore

package database

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/option"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type FirestoreTaskStore struct {
	client         *firestore.Client
	collectionName string
	projectID      string
	retryAttempts  int
	retryDelay     time.Duration
}

type FirestoreConfig struct {
	ProjectID         string
	CollectionName    string
	CredentialsFile   string
	RetryAttempts     int
	RetryInitialDelay time.Duration
}

func NewFirestoreTaskStore(ctx context.Context, config FirestoreConfig) (*FirestoreTaskStore, error) {
	var client *firestore.Client
	var err error

	opts := []option.ClientOption{}
	if config.CredentialsFile != "" {
		opts = append(opts, option.WithCredentialsFile(config.CredentialsFile))
	}

	retryAttempts := config.RetryAttempts
	if retryAttempts <= 0 {
		retryAttempts = 3
	}

	retryDelay := config.RetryInitialDelay
	if retryDelay <= 0 {
		retryDelay = 100 * time.Millisecond
	}

	for attempt := 0; attempt < retryAttempts; attempt++ {
		client, err = firestore.NewClient(ctx, config.ProjectID, opts...)
		if err == nil {
			break
		}
		if attempt < retryAttempts-1 {
			time.Sleep(retryDelay)
			retryDelay *= 2
		}
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create Firestore client: %w", err)
	}

	return &FirestoreTaskStore{
		client:         client,
		collectionName: config.CollectionName,
		projectID:      config.ProjectID,
		retryAttempts:  retryAttempts,
		retryDelay:     retryDelay,
	}, nil
}

func (f *FirestoreTaskStore) GetTask(ctx context.Context, id string) (*Todo2Task, error) {
	doc, err := f.client.Collection(f.collectionName).Doc(id).Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	var task Todo2Task
	if err := doc.DataTo(&task); err != nil {
		return nil, fmt.Errorf("failed to unmarshal task: %w", err)
	}

	return &task, nil
}

func (f *FirestoreTaskStore) UpdateTask(ctx context.Context, task *Todo2Task) error {
	task.LastModified = time.Now().Format(time.RFC3339)

	_, err := f.client.Collection(f.collectionName).Doc(task.ID).Set(ctx, task, firestore.MergeAll)
	if err != nil {
		return fmt.Errorf("failed to update task: %w", err)
	}

	return nil
}

func (f *FirestoreTaskStore) ListTasks(ctx context.Context, filters *TaskFilters) ([]*Todo2Task, error) {
	collection := f.client.Collection(f.collectionName)

	var query *firestore.Query
	if filters != nil {
		query = collection.Query
		if filters.Status != "" {
			query = query.Where("status", "==", filters.Status)
		}
		if filters.Priority != "" {
			query = query.Where("priority", "==", filters.Priority)
		}
		if filters.ProjectID != "" {
			query = query.Where("project_id", "==", filters.ProjectID)
		}
	} else {
		query = collection.Query
	}

	docs, err := query.Documents(ctx).GetAll()
	if err != nil {
		return nil, fmt.Errorf("failed to list tasks: %w", err)
	}

	tasks := make([]*Todo2Task, 0, len(docs))
	for _, doc := range docs {
		var task Todo2Task
		if err := doc.DataTo(&task); err != nil {
			continue
		}
		tasks = append(tasks, &task)
	}

	return tasks, nil
}

func (f *FirestoreTaskStore) CreateTask(ctx context.Context, task *Todo2Task) error {
	now := time.Now().Format(time.RFC3339)
	if task.CreatedAt == "" {
		task.CreatedAt = now
	}
	task.LastModified = now

	_, err := f.client.Collection(f.collectionName).Doc(task.ID).Create(ctx, task)
	if err != nil {
		return fmt.Errorf("failed to create task: %w", err)
	}

	return nil
}

func (f *FirestoreTaskStore) DeleteTask(ctx context.Context, id string) error {
	_, err := f.client.Collection(f.collectionName).Doc(id).Delete(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil
		}
		return fmt.Errorf("failed to delete task: %w", err)
	}

	return nil
}

func (f *FirestoreTaskStore) Close() error {
	return f.client.Close()
}
