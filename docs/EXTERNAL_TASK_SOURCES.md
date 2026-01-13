# External Task Source Integration

## Overview

The external task sync system allows synchronizing Todo2 tasks with various external task management services:

- **Google Tasks** - Google Workspace Tasks API
- **Any.Do** - Any.Do task management
- **Microsoft To-Do** - Microsoft Graph API
- **GitHub Issues** - GitHub Issues API
- **Agentic-Tools** - Existing MCP-based task management

## Architecture

### External Source Interface

All external sources implement the `ExternalSource` interface:

```go
type ExternalSource interface {
    Type() ExternalSourceType
    Name() string
    SyncTasks(ctx context.Context, todo2Tasks []*Todo2Task, options SyncOptions) (*SyncResult, error)
    GetTasks(ctx context.Context, options GetTasksOptions) ([]*ExternalTask, error)
    CreateTask(ctx context.Context, task *Todo2Task) (*ExternalTask, error)
    UpdateTask(ctx context.Context, externalID string, task *Todo2Task) (*ExternalTask, error)
    DeleteTask(ctx context.Context, externalID string) error
    TestConnection(ctx context.Context) error
}
```

### Status Mapping

Each source provides status mapping between Todo2 and the external system:

- **Todo2**: `Todo`, `In Progress`, `Done`, `Review`, `Cancelled`
- **External**: Varies by service (e.g., `pending`, `in-progress`, `completed`)

Default mapping is provided, but can be customized per source.

## Supported Sources

### 1. Google Tasks

**API Reference**: [Google Tasks API](https://developers.google.com/workspace/tasks/reference/rest)

**Features**:
- OAuth2 authentication
- Task lists support
- Due dates
- Notes/descriptions

**Configuration**:
```go
config := &GoogleTasksConfig{
    CredentialsPath: "path/to/credentials.json",
    TokenPath:       "path/to/token.json",
    TaskListID:      "@default", // or specific list ID
}
source, err := NewGoogleTasksSource(config)
```

**Environment Variables**:
```bash
export GOOGLE_TASKS_CREDENTIALS_PATH="/path/to/credentials.json"
export GOOGLE_TASKS_TOKEN_PATH="/path/to/token.json"
export GOOGLE_TASKS_LIST_ID="@default"
```

### 2. Any.Do

**API Reference**: [Any.Do API (GitHub)](https://github.com/davoam/anydo-api)

**Features**:
- Email/password authentication
- Categories support
- Due dates (today, tomorrow, upcoming, someday)
- Task sync

**Configuration**:
```go
config := &AnyDoConfig{
    Email:    "user@example.com",
    Password: "password",
}
source, err := NewAnyDoSource(config)
```

**Environment Variables**:
```bash
export ANYDO_EMAIL="user@example.com"
export ANYDO_PASSWORD="password"
```

### 3. Microsoft To-Do

**API Reference**: [Microsoft Graph To-Do API](https://learn.microsoft.com/en-us/graph/api/resources/todo-overview?view=graph-rest-1.0)

**Features**:
- Microsoft Graph API authentication
- Task lists support
- Due dates and reminders
- Attachments

**Configuration**:
```go
config := &MicrosoftTodoConfig{
    TenantID:     "tenant-id",
    ClientID:     "client-id",
    ClientSecret: "client-secret",
    TaskListID:   "list-id", // optional, uses default if empty
}
source, err := NewMicrosoftTodoSource(config)
```

**Environment Variables**:
```bash
export MS_TODO_TENANT_ID="tenant-id"
export MS_TODO_CLIENT_ID="client-id"
export MS_TODO_CLIENT_SECRET="client-secret"
export MS_TODO_LIST_ID="list-id"
```

### 4. GitHub Issues

**API Reference**: [GitHub Issues API](https://docs.github.com/en/rest/issues/issues?apiVersion=2022-11-28)

**Features**:
- Personal access token authentication
- Repository-specific issues
- Labels (mapped to tags)
- Assignees
- Milestones

**Configuration**:
```go
config := &GitHubIssuesConfig{
    Token:      "ghp_...",
    Owner:      "username",
    Repository: "repo-name",
}
source, err := NewGitHubIssuesSource(config)
```

**Environment Variables**:
```bash
export GITHUB_TOKEN="ghp_..."
export GITHUB_OWNER="username"
export GITHUB_REPO="repo-name"
```

## Usage

### Basic Sync

```go
import "github.com/davidl71/exarp-go/internal/tasksync"

// Get external source
source, err := tasksync.GetExternalSource(tasksync.SourceGoogleTasks)
if err != nil {
    return err
}

// Load Todo2 tasks
todo2Tasks, err := LoadTodo2Tasks(projectRoot)
if err != nil {
    return err
}

// Sync options
options := tasksync.SyncOptions{
    DryRun:              false,
    Direction:           tasksync.DirectionBidirectional,
    ConflictResolution:  tasksync.ConflictResolutionNewerWins,
}

// Perform sync
result, err := source.SyncTasks(ctx, todo2Tasks, options)
if err != nil {
    return err
}

// Process results
fmt.Printf("Matches: %d\n", result.Stats.MatchesFound)
fmt.Printf("Conflicts: %d\n", result.Stats.ConflictsDetected)
fmt.Printf("New tasks: %d\n", result.Stats.NewTodo2Tasks + result.Stats.NewExternalTasks)
```

### Via Task Workflow Tool

```bash
# Sync with Google Tasks
exarp-go -tool task_workflow -args '{
  "action": "sync",
  "external": true,
  "external_source": "google-tasks",
  "dry_run": false
}'

# Sync with GitHub Issues
exarp-go -tool task_workflow -args '{
  "action": "sync",
  "external": true,
  "external_source": "github-issues",
  "dry_run": false
}'
```

## Sync Directions

- **Bidirectional**: Sync both ways (default)
- **Todo2 to External**: Only push Todo2 tasks to external source
- **External to Todo2**: Only pull tasks from external source

## Conflict Resolution

- **Todo2 Wins**: Todo2 version takes precedence
- **External Wins**: External version takes precedence
- **Newer Wins**: Most recently modified version wins
- **Manual**: Conflicts are reported but not auto-resolved

## Status Mapping

Default status mapping:

| Todo2 | Google Tasks | Any.Do | Microsoft To-Do | GitHub Issues |
|-------|-------------|--------|-----------------|---------------|
| Todo | needsAction | pending | notStarted | open |
| In Progress | needsAction | in-progress | inProgress | open |
| Done | completed | completed | completed | closed |
| Review | needsAction | in-progress | inProgress | open |
| Cancelled | needsAction | cancelled | completed | closed |

Custom mappings can be provided per source.

## Authentication

### Google Tasks
1. Create OAuth2 credentials in Google Cloud Console
2. Download credentials JSON file
3. First run will open browser for OAuth flow
4. Token is saved for subsequent runs

### Microsoft To-Do
1. Register app in Azure AD
2. Get tenant ID, client ID, and client secret
3. Grant Tasks.ReadWrite permission

### GitHub Issues
1. Create personal access token with `repo` scope
2. Use token for authentication

### Any.Do
1. Use email and password
2. Token is obtained via login API

## Error Handling

Sync operations return detailed error information:

```go
for _, syncErr := range result.Errors {
    fmt.Printf("Error on task %s: %v\n", syncErr.TaskID, syncErr.Error)
}
```

## Rate Limiting

External APIs may have rate limits. The sync system:
- Respects rate limits
- Implements exponential backoff
- Reports rate limit errors

## Future Enhancements

- [ ] Two-way sync with conflict resolution UI
- [ ] Scheduled automatic sync
- [ ] Webhook support for real-time sync
- [ ] Additional sources (Asana, Trello, Jira, etc.)
- [ ] Batch operations for better performance
- [ ] Sync history and audit log
