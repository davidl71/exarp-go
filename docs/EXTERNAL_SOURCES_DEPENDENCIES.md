# External Task Sources - Dependencies

## Required Dependencies

To use external task source synchronization, you need to install the following official client libraries:

### Google Tasks

```bash
go get google.golang.org/api/tasks/v1
go get golang.org/x/oauth2/google
```

**Already in go.mod**: `golang.org/x/oauth2` ✅

**Status**: ✅ Implementation complete using official library

### Microsoft To-Do (Graph API)

```bash
go get github.com/microsoftgraph/msgraph-sdk-go
go get github.com/microsoftgraph/msgraph-sdk-go/models
```

**Note**: The Microsoft Graph SDK API may differ from the implementation. The SDK uses a builder pattern and may require adjustments.

**Status**: ⚠️ Implementation created, may need API adjustments

### GitHub Issues

```bash
go get github.com/google/go-github/v60/github
```

**Already in go.mod**: `golang.org/x/oauth2` ✅

**Status**: ✅ Implementation complete using official library

### Any.Do

**Status**: ❌ No official library - custom HTTP client needed

## Installation

### All at once:

```bash
# Google Tasks
go get google.golang.org/api/tasks/v1
go get golang.org/x/oauth2/google

# Microsoft To-Do
go get github.com/microsoftgraph/msgraph-sdk-go

# GitHub Issues
go get github.com/google/go-github/v60/github
```

### Or add to go.mod manually:

```go
require (
    // ... existing dependencies ...
    google.golang.org/api/tasks/v1 v0.0.0-...
    github.com/microsoftgraph/msgraph-sdk-go v0.0.0-...
    github.com/google/go-github/v60 v60.0.0
)
```

## Optional Dependencies

These are already included:
- `golang.org/x/oauth2` - OAuth2 support (already in go.mod)
- `golang.org/x/oauth2/google` - Google OAuth2 (needs to be added)

## Build Tags

No special build tags required for these libraries. They work with standard Go builds.

## Testing

After installing dependencies, you can test each source:

```go
// Test Google Tasks
source, err := tasksync.NewGoogleTasksSource(&tasksync.GoogleTasksConfig{
    CredentialsPath: "credentials.json",
    TokenPath:       "token.json",
})
if err != nil {
    log.Fatal(err)
}

err = source.TestConnection(context.Background())
if err != nil {
    log.Fatal(err)
}
```

## Troubleshooting

### Microsoft Graph SDK

The Microsoft Graph SDK uses a builder pattern. If you encounter API errors, check:
- SDK version compatibility
- Authentication method (client credentials vs. delegated permissions)
- API endpoint URLs

### GitHub API

- Ensure token has `repo` scope for private repos
- For public repos, token is optional but recommended for rate limits

### Google Tasks

- OAuth2 flow requires browser interaction on first run
- Token is cached for subsequent runs
- Ensure credentials.json has Tasks API enabled
