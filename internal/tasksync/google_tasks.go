package tasksync

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/davidl71/exarp-go/internal/models"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/tasks/v1"
)

// Note: This implementation uses the official Google Tasks API client library
// Install with: go get google.golang.org/api/tasks/v1
// Reference: https://developers.google.com/workspace/tasks/quickstart/go

// GoogleTasksSource implements ExternalSource for Google Tasks API
// Reference: https://developers.google.com/workspace/tasks/reference/rest
type GoogleTasksSource struct {
	service *tasks.Service
	config  *GoogleTasksConfig
}

// GoogleTasksConfig contains configuration for Google Tasks API
type GoogleTasksConfig struct {
	// CredentialsPath is the path to OAuth2 credentials JSON file
	CredentialsPath string

	// TokenPath is the path to store/load OAuth2 token
	TokenPath string

	// TaskListID is the Google Tasks list ID to sync with
	// Use "@default" for the default task list
	TaskListID string

	// ClientID and ClientSecret for OAuth2 (alternative to CredentialsPath)
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

// NewGoogleTasksSource creates a new Google Tasks source
func NewGoogleTasksSource(config *GoogleTasksConfig) (*GoogleTasksSource, error) {
	ctx := context.Background()

	// Load OAuth2 config
	var oauthConfig *oauth2.Config
	if config.CredentialsPath != "" {
		// Load from credentials file
		b, err := os.ReadFile(config.CredentialsPath)
		if err != nil {
			return nil, fmt.Errorf("unable to read credentials file: %w", err)
		}

		config, err := google.ConfigFromJSON(b, tasks.TasksScope)
		if err != nil {
			return nil, fmt.Errorf("unable to parse credentials: %w", err)
		}
		oauthConfig = config
	} else if config.ClientID != "" && config.ClientSecret != "" {
		// Use provided client credentials
		oauthConfig = &oauth2.Config{
			ClientID:     config.ClientID,
			ClientSecret: config.ClientSecret,
			RedirectURL:  config.RedirectURL,
			Scopes:       []string{tasks.TasksScope},
			Endpoint:     google.Endpoint,
		}
	} else {
		return nil, fmt.Errorf("either CredentialsPath or ClientID/ClientSecret must be provided")
	}

	// Load or get token
	token, err := getTokenFromFile(config.TokenPath)
	if err != nil {
		// Need to get new token
		token, err = getTokenFromWeb(oauthConfig)
		if err != nil {
			return nil, fmt.Errorf("unable to get token: %w", err)
		}
		saveToken(config.TokenPath, token)
	}

	// Create service
	client := oauthConfig.Client(ctx, token)
	service, err := tasks.New(client)
	if err != nil {
		return nil, fmt.Errorf("unable to create tasks service: %w", err)
	}

	return &GoogleTasksSource{
		service: service,
		config:  config,
	}, nil
}

// Type returns the source type
func (s *GoogleTasksSource) Type() ExternalSourceType {
	return SourceGoogleTasks
}

// Name returns a human-readable name
func (s *GoogleTasksSource) Name() string {
	return "Google Tasks"
}

// SyncTasks synchronizes tasks between Todo2 and Google Tasks
func (s *GoogleTasksSource) SyncTasks(ctx context.Context, todo2Tasks []*models.Todo2Task, options SyncOptions) (*SyncResult, error) {
	result := &SyncResult{
		Matches:          []TaskMatch{},
		Conflicts:        []TaskConflict{},
		NewTodo2Tasks:    []*models.Todo2Task{},
		NewExternalTasks: []*ExternalTask{},
		UpdatedTasks:     []TaskUpdate{},
		Errors:           []SyncError{},
	}

	// Get tasks from Google Tasks
	externalTasks, err := s.GetTasks(ctx, GetTasksOptions{
		IncludeCompleted: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get Google Tasks: %w", err)
	}

	// Build mapping by title/content for matching
	todo2Map := make(map[string]*models.Todo2Task)
	for _, task := range todo2Tasks {
		key := fmt.Sprintf("%s:%s", task.Content, task.ID)
		todo2Map[key] = task
	}

	externalMap := make(map[string]*ExternalTask)
	for _, task := range externalTasks {
		key := fmt.Sprintf("%s:%s", task.Title, task.ExternalID)
		externalMap[key] = task
	}

	// Match tasks
	for key, todo2Task := range todo2Map {
		if extTask, exists := externalMap[key]; exists {
			// Check if they match
			if tasksMatch(todo2Task, extTask) {
				result.Matches = append(result.Matches, TaskMatch{
					Todo2Task:    todo2Task,
					ExternalTask: extTask,
					MatchScore:   1.0,
				})
			} else {
				result.Conflicts = append(result.Conflicts, TaskConflict{
					Todo2Task:    todo2Task,
					ExternalTask: extTask,
					Differences:  getDifferences(todo2Task, extTask),
				})
			}
		} else if options.Direction == DirectionBidirectional || options.Direction == DirectionTodo2ToExternal {
			// Create in Google Tasks
			if !options.DryRun {
				extTask, err := s.CreateTask(ctx, todo2Task)
				if err != nil {
					result.Errors = append(result.Errors, SyncError{
						TaskID:    todo2Task.ID,
						Source:    SourceGoogleTasks,
						Operation: "create",
						Error:     err,
					})
					continue
				}
				result.NewExternalTasks = append(result.NewExternalTasks, extTask)
			}
		}
	}

	// Find tasks in Google Tasks that don't exist in Todo2
	for key, extTask := range externalMap {
		if _, exists := todo2Map[key]; !exists {
			if options.Direction == DirectionBidirectional || options.Direction == DirectionExternalToTodo2 {
				// Create in Todo2
				todo2Task := s.externalToTodo2(extTask)
				result.NewTodo2Tasks = append(result.NewTodo2Tasks, todo2Task)
			}
		}
	}

	// Update stats
	result.Stats = SyncStats{
		TotalTodo2Tasks:    len(todo2Tasks),
		TotalExternalTasks: len(externalTasks),
		MatchesFound:       len(result.Matches),
		ConflictsDetected:  len(result.Conflicts),
		NewTodo2Tasks:      len(result.NewTodo2Tasks),
		NewExternalTasks:    len(result.NewExternalTasks),
		UpdatedTasks:        len(result.UpdatedTasks),
		Errors:              len(result.Errors),
	}

	return result, nil
}

// GetTasks retrieves tasks from Google Tasks
func (s *GoogleTasksSource) GetTasks(ctx context.Context, options GetTasksOptions) ([]*ExternalTask, error) {
	taskListID := s.config.TaskListID
	if taskListID == "" {
		taskListID = "@default"
	}

	call := s.service.Tasks.List(taskListID)
	if options.Status != "" {
		// Google Tasks doesn't have status filter, we'll filter after
	}
	if !options.IncludeCompleted {
		call = call.ShowCompleted(false)
	}
	if options.Limit > 0 {
		call = call.MaxResults(int64(options.Limit))
	}

	resp, err := call.Do()
	if err != nil {
		return nil, fmt.Errorf("failed to list tasks: %w", err)
	}

	var externalTasks []*ExternalTask
	for _, item := range resp.Items {
		extTask := s.googleTaskToExternal(item)
		
		// Apply filters
		if options.Status != "" && extTask.Status != options.Status {
			continue
		}
		if options.Since != nil && extTask.LastModified.Before(*options.Since) {
			continue
		}

		externalTasks = append(externalTasks, extTask)
	}

	return externalTasks, nil
}

// CreateTask creates a task in Google Tasks
func (s *GoogleTasksSource) CreateTask(ctx context.Context, task *models.Todo2Task) (*ExternalTask, error) {
	taskListID := s.config.TaskListID
	if taskListID == "" {
		taskListID = "@default"
	}

	googleTask := s.todo2ToGoogleTask(task)
	created, err := s.service.Tasks.Insert(taskListID, googleTask).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to create task: %w", err)
	}

	return s.googleTaskToExternal(created), nil
}

// UpdateTask updates a task in Google Tasks
func (s *GoogleTasksSource) UpdateTask(ctx context.Context, externalID string, task *models.Todo2Task) (*ExternalTask, error) {
	taskListID := s.config.TaskListID
	if taskListID == "" {
		taskListID = "@default"
	}

	googleTask := s.todo2ToGoogleTask(task)
	googleTask.Id = externalID
	updated, err := s.service.Tasks.Update(taskListID, externalID, googleTask).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to update task: %w", err)
	}

	return s.googleTaskToExternal(updated), nil
}

// DeleteTask deletes a task from Google Tasks
func (s *GoogleTasksSource) DeleteTask(ctx context.Context, externalID string) error {
	taskListID := s.config.TaskListID
	if taskListID == "" {
		taskListID = "@default"
	}

	return s.service.Tasks.Delete(taskListID, externalID).Do()
}

// TestConnection tests the connection to Google Tasks
func (s *GoogleTasksSource) TestConnection(ctx context.Context) error {
	_, err := s.service.Tasklists.List().MaxResults(1).Do()
	return err
}

// Helper functions

func (s *GoogleTasksSource) googleTaskToExternal(gt *tasks.Task) *ExternalTask {
	extTask := &ExternalTask{
		ExternalID:   gt.Id,
		Title:        gt.Title,
		Description:  gt.Notes,
		Status:       gt.Status, // "needsAction" or "completed"
		Completed:    gt.Status == "completed",
		Source:       SourceGoogleTasks,
		Metadata:     make(map[string]interface{}),
	}

	// Parse due date
	if gt.Due != "" {
		if dueDate, err := time.Parse(time.RFC3339, gt.Due); err == nil {
			extTask.DueDate = &dueDate
		}
	}

	// Map status
	mapper := NewDefaultStatusMapper()
	if gt.Status == "completed" {
		extTask.Status = mapper.ExternalToTodo2("completed")
	} else {
		extTask.Status = mapper.ExternalToTodo2("pending")
	}

	// Last modified
	if gt.Updated != "" {
		if updated, err := time.Parse(time.RFC3339, gt.Updated); err == nil {
			extTask.LastModified = updated
		}
	}

	return extTask
}

func (s *GoogleTasksSource) todo2ToGoogleTask(task *models.Todo2Task) *tasks.Task {
	googleTask := &tasks.Task{
		Title: task.Content,
		Notes: task.LongDescription,
	}

	// Map status
	mapper := NewDefaultStatusMapper()
	externalStatus := mapper.Todo2ToExternal(task.Status)
	if externalStatus == "completed" || task.Status == "Done" {
		googleTask.Status = "completed"
	} else {
		googleTask.Status = "needsAction"
	}

	// Due date
	if task.Metadata != nil {
		if dueDateStr, ok := task.Metadata["due_date"].(string); ok {
			googleTask.Due = dueDateStr
		}
	}

	return googleTask
}

func (s *GoogleTasksSource) externalToTodo2(extTask *ExternalTask) *models.Todo2Task {
	task := &models.Todo2Task{
		ID:             fmt.Sprintf("GT-%s", extTask.ExternalID),
		Content:        extTask.Title,
		LongDescription: extTask.Description,
		Status:         extTask.Status,
		Priority:       extTask.Priority,
		Tags:           extTask.Tags,
		Metadata:       extTask.Metadata,
	}

	if extTask.DueDate != nil {
		if task.Metadata == nil {
			task.Metadata = make(map[string]interface{})
		}
		task.Metadata["due_date"] = extTask.DueDate.Format(time.RFC3339)
	}

	return task
}

// OAuth2 helper functions

func getTokenFromFile(filepath string) (*oauth2.Token, error) {
	if filepath == "" {
		return nil, fmt.Errorf("token file path not provided")
	}
	
	f, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

func getTokenFromWeb(config *oauth2.Config) (*oauth2.Token, error) {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		return nil, fmt.Errorf("unable to read authorization code: %w", err)
	}

	tok, err := config.Exchange(context.Background(), authCode)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve token from web: %w", err)
	}
	return tok, nil
}

func saveToken(filepath string, token *oauth2.Token) error {
	if filepath == "" {
		return fmt.Errorf("token file path not provided")
	}
	
	f, err := os.OpenFile(filepath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("unable to cache oauth token: %w", err)
	}
	defer f.Close()
	
	return json.NewEncoder(f).Encode(token)
}

func tasksMatch(todo2Task *models.Todo2Task, extTask *ExternalTask) bool {
	// Simple matching by title/content
	return todo2Task.Content == extTask.Title
}

func getDifferences(todo2Task *models.Todo2Task, extTask *ExternalTask) []string {
	var diffs []string
	if todo2Task.Content != extTask.Title {
		diffs = append(diffs, "title")
	}
	if todo2Task.Status != extTask.Status {
		diffs = append(diffs, "status")
	}
	// Add more field comparisons
	return diffs
}
