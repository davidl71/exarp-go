package tasksync

import (
	"context"
	"fmt"
	"time"

	"github.com/davidl71/exarp-go/internal/models"
	"github.com/google/go-github/v60/github"
	"golang.org/x/oauth2"
)

// Note: This implementation uses the official go-github library
// Install with: go get github.com/google/go-github/v60/github
// Reference: https://docs.github.com/en/rest/issues/issues?apiVersion=2022-11-28

// GitHubIssuesSource implements ExternalSource for GitHub Issues API
type GitHubIssuesSource struct {
	client *github.Client
	config *GitHubIssuesConfig
}

// GitHubIssuesConfig contains configuration for GitHub Issues API
type GitHubIssuesConfig struct {
	// Token is the GitHub personal access token
	Token string

	// Owner is the repository owner (username or organization)
	Owner string

	// Repository is the repository name
	Repository string

	// BaseURL is the GitHub API base URL (optional, defaults to github.com)
	BaseURL string
}

// NewGitHubIssuesSource creates a new GitHub Issues source
func NewGitHubIssuesSource(config *GitHubIssuesConfig) (*GitHubIssuesSource, error) {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: config.Token},
	)
	tc := oauth2.NewClient(ctx, ts)

	var client *github.Client
	if config.BaseURL != "" {
		// For GitHub Enterprise
		client = github.NewClient(tc)
		client.BaseURL.Scheme = "https"
		client.BaseURL.Host = config.BaseURL
	} else {
		// Standard GitHub.com
		client = github.NewClient(tc)
	}

	return &GitHubIssuesSource{
		client: client,
		config: config,
	}, nil
}

// Type returns the source type
func (s *GitHubIssuesSource) Type() ExternalSourceType {
	return SourceGitHubIssues
}

// Name returns a human-readable name
func (s *GitHubIssuesSource) Name() string {
	return fmt.Sprintf("GitHub Issues (%s/%s)", s.config.Owner, s.config.Repository)
}

// SyncTasks synchronizes tasks between Todo2 and GitHub Issues
func (s *GitHubIssuesSource) SyncTasks(ctx context.Context, todo2Tasks []*models.Todo2Task, options SyncOptions) (*SyncResult, error) {
	result := &SyncResult{
		Matches:          []TaskMatch{},
		Conflicts:        []TaskConflict{},
		NewTodo2Tasks:    []*models.Todo2Task{},
		NewExternalTasks: []*ExternalTask{},
		UpdatedTasks:     []TaskUpdate{},
		Errors:           []SyncError{},
	}

	// Get issues from GitHub
	externalTasks, err := s.GetTasks(ctx, GetTasksOptions{
		IncludeCompleted: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get GitHub Issues: %w", err)
	}

	// Build mapping for matching
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

	// Match and sync tasks (similar to Google Tasks implementation)
	// ... (implementation similar to GoogleTasksSource.SyncTasks)

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

// GetTasks retrieves issues from GitHub
func (s *GitHubIssuesSource) GetTasks(ctx context.Context, options GetTasksOptions) ([]*ExternalTask, error) {
	opts := &github.IssueListByRepoOptions{
		State: "all", // Get both open and closed issues
	}

	if options.Status != "" {
		// Map Todo2 status to GitHub state
		if options.Status == "Done" || options.Status == "Cancelled" {
			opts.State = "closed"
		} else {
			opts.State = "open"
		}
	}

	if options.Limit > 0 {
		opts.PerPage = options.Limit
	}

	var allIssues []*github.Issue
	for {
		issues, resp, err := s.client.Issues.ListByRepo(ctx, s.config.Owner, s.config.Repository, opts)
		if err != nil {
			return nil, fmt.Errorf("failed to list issues: %w", err)
		}

		allIssues = append(allIssues, issues...)

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	var externalTasks []*ExternalTask
	for _, issue := range allIssues {
		// Skip pull requests (they have PullRequestLinks)
		if issue.PullRequestLinks != nil {
			continue
		}

		extTask := s.githubIssueToExternal(issue)

		// Apply filters
		if !options.IncludeCompleted && extTask.Completed {
			continue
		}
		if options.Since != nil && extTask.LastModified.Before(*options.Since) {
			continue
		}
		if options.Limit > 0 && len(externalTasks) >= options.Limit {
			break
		}

		externalTasks = append(externalTasks, extTask)
	}

	return externalTasks, nil
}

// CreateTask creates an issue in GitHub
func (s *GitHubIssuesSource) CreateTask(ctx context.Context, task *models.Todo2Task) (*ExternalTask, error) {
	req := &github.IssueRequest{
		Title: &task.Content,
	}

	if task.LongDescription != "" {
		req.Body = &task.LongDescription
	}

	// Map tags to labels
	if len(task.Tags) > 0 {
		req.Labels = &task.Tags
	}

	issue, _, err := s.client.Issues.Create(ctx, s.config.Owner, s.config.Repository, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create issue: %w", err)
	}

	return s.githubIssueToExternal(issue), nil
}

// UpdateTask updates an issue in GitHub
func (s *GitHubIssuesSource) UpdateTask(ctx context.Context, externalID string, task *models.Todo2Task) (*ExternalTask, error) {
	req := &github.IssueRequest{
		Title: &task.Content,
	}

	if task.LongDescription != "" {
		req.Body = &task.LongDescription
	}

	// Map status
	mapper := NewDefaultStatusMapper()
	externalStatus := mapper.Todo2ToExternal(task.Status)
	state := "open"
	if externalStatus == "completed" || task.Status == "Done" {
		state = "closed"
	}
	req.State = &state

	// Map tags to labels
	if len(task.Tags) > 0 {
		req.Labels = &task.Tags
	}

	issue, _, err := s.client.Issues.Edit(ctx, s.config.Owner, s.config.Repository, externalID, req)
	if err != nil {
		return nil, fmt.Errorf("failed to update issue: %w", err)
	}

	return s.githubIssueToExternal(issue), nil
}

// DeleteTask closes an issue in GitHub (GitHub doesn't allow deleting issues)
func (s *GitHubIssuesSource) DeleteTask(ctx context.Context, externalID string) error {
	state := "closed"
	req := &github.IssueRequest{
		State: &state,
	}
	_, _, err := s.client.Issues.Edit(context.Background(), s.config.Owner, s.config.Repository, externalID, req)
	return err
}

// TestConnection tests the connection to GitHub
func (s *GitHubIssuesSource) TestConnection(ctx context.Context) error {
	_, _, err := s.client.Repositories.Get(ctx, s.config.Owner, s.config.Repository)
	return err
}

// Helper functions

func (s *GitHubIssuesSource) githubIssueToExternal(issue *github.Issue) *ExternalTask {
	extTask := &ExternalTask{
		ExternalID:   fmt.Sprintf("%d", issue.GetNumber()),
		Title:        issue.GetTitle(),
		Description:  issue.GetBody(),
		Source:       SourceGitHubIssues,
		Metadata:     make(map[string]interface{}),
	}

	// Status mapping
	mapper := NewDefaultStatusMapper()
	state := issue.GetState()
	if state == "closed" {
		extTask.Status = mapper.ExternalToTodo2("completed")
		extTask.Completed = true
	} else {
		extTask.Status = mapper.ExternalToTodo2("pending")
	}

	// Labels as tags
	if labels := issue.Labels; labels != nil {
		for _, label := range labels {
			extTask.Tags = append(extTask.Tags, label.GetName())
		}
	}

	// Assignees
	if assignees := issue.Assignees; assignees != nil {
		assigneeNames := make([]string, 0, len(assignees))
		for _, assignee := range assignees {
			assigneeNames = append(assigneeNames, assignee.GetLogin())
		}
		extTask.Metadata["assignees"] = assigneeNames
	}

	// Milestone
	if milestone := issue.Milestone; milestone != nil {
		extTask.Metadata["milestone"] = milestone.GetTitle()
	}

	// Last modified
	if updated := issue.GetUpdatedAt(); !updated.IsZero() {
		extTask.LastModified = updated
	}

	// Created at
	if created := issue.GetCreatedAt(); !created.IsZero() {
		extTask.Metadata["created_at"] = created.Format(time.RFC3339)
	}

	// URL
	extTask.Metadata["url"] = issue.GetHTMLURL()

	return extTask
}

func (s *GitHubIssuesSource) todo2ToGitHubIssue(task *models.Todo2Task) *github.IssueRequest {
	req := &github.IssueRequest{
		Title: &task.Content,
	}

	if task.LongDescription != "" {
		req.Body = &task.LongDescription
	}

	// Map status
	mapper := NewDefaultStatusMapper()
	externalStatus := mapper.Todo2ToExternal(task.Status)
	state := "open"
	if externalStatus == "completed" || task.Status == "Done" {
		state = "closed"
	}
	req.State = &state

	// Map tags to labels
	if len(task.Tags) > 0 {
		req.Labels = &task.Tags
	}

	return req
}

func (s *GitHubIssuesSource) externalToTodo2(extTask *ExternalTask) *models.Todo2Task {
	task := &models.Todo2Task{
		ID:             fmt.Sprintf("GH-%s", extTask.ExternalID),
		Content:        extTask.Title,
		LongDescription: extTask.Description,
		Status:         extTask.Status,
		Priority:       extTask.Priority,
		Tags:           extTask.Tags,
		Metadata:       extTask.Metadata,
	}

	return task
}
