# External Task Sources - Implementation Status

## Summary

Implemented external task source synchronization using **official client libraries** where available.

## ✅ Completed Implementations

### 1. Google Tasks ✅

**Library**: `google.golang.org/api/tasks/v1` (Official)

**Status**: ✅ **Complete** - Uses official Google client library

**Features**:
- OAuth2 authentication (credentials file or client ID/secret)
- Full CRUD operations
- Task list support
- Status mapping (Todo2 ↔ Google Tasks)
- Due dates and notes

**Implementation**: `internal/tasksync/google_tasks.go`

**Dependencies**:
```bash
go get google.golang.org/api/tasks/v1
go get golang.org/x/oauth2/google
```

### 2. GitHub Issues ✅

**Library**: `github.com/google/go-github/v60` (Official)

**Status**: ✅ **Complete** - Uses official go-github library

**Features**:
- Personal access token authentication
- Repository-specific issues
- Labels (mapped to tags)
- Assignees and milestones
- Status mapping (open/closed ↔ Todo2 status)
- Issue URLs in metadata

**Implementation**: `internal/tasksync/github_issues.go`

**Dependencies**:
```bash
go get github.com/google/go-github/v60/github
```

## ⚠️ Partial Implementations

### 3. Microsoft To-Do ⚠️

**Library**: `github.com/microsoftgraph/msgraph-sdk-go` (Official)

**Status**: ⚠️ **Structure Complete, API Pending** - SDK API needs verification

**Features** (Planned):
- OAuth2 client credentials flow
- Task list support
- Due dates and reminders
- Status mapping

**Implementation**: `internal/tasksync/microsoft_todo.go`

**Note**: The Microsoft Graph SDK uses a builder pattern that may differ from the implementation. The structure is in place, but API calls need to be verified once the SDK is added.

**Dependencies**:
```bash
go get github.com/microsoftgraph/msgraph-sdk-go
```

**Next Steps**:
1. Add SDK to go.mod
2. Verify SDK API structure
3. Complete implementation based on actual API

## ❌ Not Yet Implemented

### 4. Any.Do ❌

**Status**: ❌ **No Official Library** - Custom HTTP client needed

**Options**:
- Use [Node.js library](https://github.com/davoam/anydo-api) as reference
- Build custom HTTP client
- May need to reverse-engineer API

## Architecture

All implementations follow the `ExternalSource` interface:

```go
type ExternalSource interface {
    Type() ExternalSourceType
    Name() string
    SyncTasks(ctx, todo2Tasks, options) (*SyncResult, error)
    GetTasks(ctx, options) ([]*ExternalTask, error)
    CreateTask(ctx, task) (*ExternalTask, error)
    UpdateTask(ctx, externalID, task) (*ExternalTask, error)
    DeleteTask(ctx, externalID) error
    TestConnection(ctx) error
}
```

## Status Mapping

All sources use the `DefaultStatusMapper` for consistent status conversion:

| Todo2 | Google Tasks | GitHub Issues | Microsoft To-Do |
|-------|-------------|---------------|-----------------|
| Todo | needsAction | open | notStarted |
| In Progress | needsAction | open | inProgress |
| Done | completed | closed | completed |
| Review | needsAction | open | inProgress |
| Cancelled | needsAction | closed | completed |

## Dependencies Summary

### Required (for Google Tasks)
```bash
go get google.golang.org/api/tasks/v1
go get golang.org/x/oauth2/google
```

### Required (for GitHub Issues)
```bash
go get github.com/google/go-github/v60/github
```

### Required (for Microsoft To-Do - when completed)
```bash
go get github.com/microsoftgraph/msgraph-sdk-go
```

### Already in go.mod
- ✅ `golang.org/x/oauth2` - OAuth2 support

## Next Steps

1. **Add dependencies to go.mod**:
   ```bash
   go get google.golang.org/api/tasks/v1
   go get golang.org/x/oauth2/google
   go get github.com/google/go-github/v60/github
   ```

2. **Complete Microsoft To-Do**:
   - Add SDK: `go get github.com/microsoftgraph/msgraph-sdk-go`
   - Verify SDK API structure
   - Complete implementation

3. **Integrate with task_workflow**:
   - Update sync handler to support `external_source` parameter
   - Add source selection logic
   - Register sources in global registry

4. **Add configuration**:
   - Environment variable support
   - Config file support
   - Source-specific settings

5. **Testing**:
   - Unit tests for each source
   - Integration tests with real APIs
   - OAuth flow testing

## Usage Example (Once Integrated)

```go
// Get Google Tasks source
source, err := tasksync.GetExternalSource(tasksync.SourceGoogleTasks)
if err != nil {
    return err
}

// Sync tasks
result, err := source.SyncTasks(ctx, todo2Tasks, tasksync.SyncOptions{
    DryRun:             false,
    Direction:          tasksync.DirectionBidirectional,
    ConflictResolution: tasksync.ConflictResolutionNewerWins,
})
```

## References

- [Google Tasks API Go Quickstart](https://developers.google.com/workspace/tasks/quickstart/go)
- [go-github Library](https://github.com/google/go-github)
- [Microsoft Graph Go SDK](https://github.com/microsoftgraph/msgraph-sdk-go)
- [Microsoft Graph To-Do API](https://learn.microsoft.com/en-us/graph/api/resources/todo-overview?view=graph-rest-1.0)
- [GitHub Issues API](https://docs.github.com/en/rest/issues/issues?apiVersion=2022-11-28)
