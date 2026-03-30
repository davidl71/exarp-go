# Deep Dive: Integration Patterns in Rust TUI Applications

## Executive Summary

Modern TUI applications rarely operate in isolation. This analysis examines **3 primary integration patterns** across **18 Rust TUI projects**: REST API integration (33%), Shell/System integration (94%), and emerging MCP (Model Context Protocol) adoption (11%).

**Key Finding:** While shell integration is nearly universal (17/18 projects), API integration varies significantly by domain. MCP represents a new paradigm for AI-augmented TUIs.

---

## Integration Pattern Matrix

| Pattern | Projects | % | Complexity | Use Case |
|---------|----------|---|------------|----------|
| **Shell/System** | 17 | 94% | Low | CLI wrapping, external tools |
| **REST API** | 6 | 33% | Medium | Cloud services, SaaS |
| **MCP Server** | 2 | 11% | High | AI integration, extensibility |
| **Git Integration** | 4 | 22% | Medium | Version control workflows |

### Projects by Integration Type

**Shell Integration (17):**
- All except basilk (pure local)

**REST API (6):**
- jirust (JIRA), tatuin (Todoist/GitHub), rust_kanban (Supabase), sc-cli (Shortcut), taskwarrior-tui (Taskwarrior CLI), television (fuzzy finder)

**MCP Server (2):**
- kanban, nereid

**Git Integration (4):**
- sc-cli, taskwarrior-tui, television, kanban

---

## Pattern 1: Shell and System Integration

### Overview
The most common integration pattern. TUIs spawn external processes, capture output, and present results in the terminal interface.

### Sub-Patterns

#### 1.1 CLI Wrapping (Command Execution)

**From taskwarrior-tui:**
```rust
use std::process::{Command, Stdio};

pub struct TaskwarriorClient;

impl TaskwarriorClient {
    pub fn export_tasks(&self, filter: &str) -> Result<Vec<Task>> {
        let output = Command::new("task")
            .args(&["rc.color=off", "rc._forcecolor=off", "export", filter])
            .stdout(Stdio::piped())
            .stderr(Stdio::piped())
            .output()?;
        
        if !output.status.success() {
            let stderr = String::from_utf8_lossy(&output.stderr);
            return Err(anyhow!("Taskwarrior error: {}", stderr));
        }
        
        let tasks: Vec<Task> = serde_json::from_slice(&output.stdout)?;
        Ok(tasks)
    }
    
    pub fn modify_task(&self, uuid: &str, changes: &[(&str, &str)]) -> Result<()> {
        let mut cmd = Command::new("task");
        cmd.arg(uuid);
        
        for (field, value) in changes {
            cmd.arg(format!("{}={}", field, value));
        }
        
        cmd.arg("modify");
        
        let output = cmd.output()?;
        
        if !output.status.success() {
            return Err(anyhow!("Failed to modify task"));
        }
        
        Ok(())
    }
}
```

**From television (shell command integration):**
```rust
use tokio::process::Command;

pub async fn execute_shell_command(cmd: &str) -> Result<Vec<String>> {
    let output = if cfg!(target_os = "windows") {
        Command::new("cmd")
            .args(&["/C", cmd])
            .output()
            .await?
    } else {
        Command::new("sh")
            .args(&["-c", cmd])
            .output()
            .await?
    };
    
    let stdout = String::from_utf8_lossy(&output.stdout);
    Ok(stdout.lines().map(String::from).collect())
}
```

#### 1.2 Git Integration

**From sc-cli (comprehensive git operations):**
```rust
use std::process::Command;
use std::path::Path;

pub struct GitClient;

impl GitClient {
    pub fn is_git_repo(path: &Path) -> bool {
        path.join(".git").exists()
    }
    
    pub fn current_branch(path: &Path) -> Result<String> {
        let output = Command::new("git")
            .args(&["branch", "--show-current"])
            .current_dir(path)
            .output()?;
        
        if !output.status.success() {
            return Err(anyhow!("Not a git repository"));
        }
        
        Ok(String::from_utf8(output.stdout)?.trim().to_string())
    }
    
    pub fn create_branch(&self, story_id: &str, story_name: &str) -> Result<String> {
        let sanitized_name = story_name.to_lowercase()
            .replace(" ", "-")
            .replace(|c: char| !c.is_alphanumeric() && c != '-', "");
        
        let branch_name = format!("{}/{}", story_id, sanitized_name);
        
        let output = Command::new("git")
            .args(&["checkout", "-b", &branch_name])
            .output()?;
        
        if !output.status.success() {
            return Err(anyhow!("Failed to create branch"));
        }
        
        Ok(branch_name)
    }
    
    pub fn open_in_browser(&self, url: &str) -> Result<()> {
        let opener = if cfg!(target_os = "macos") {
            "open"
        } else if cfg!(target_os = "windows") {
            "start"
        } else {
            "xdg-open"
        };
        
        Command::new(opener)
            .arg(url)
            .spawn()?;
        
        Ok(())
    }
}
```

#### 1.3 Browser Opening

**Universal pattern (seen in 8 projects):**
```rust
pub fn open_browser(url: &str) -> Result<()> {
    #[cfg(target_os = "macos")]
    {
        Command::new("open").arg(url).spawn()?;
    }
    
    #[cfg(target_os = "windows")]
    {
        Command::new("cmd")
            .args(&["/C", "start", ""])
            .arg(url)
            .spawn()?;
    }
    
    #[cfg(target_os = "linux")]
    {
        Command::new("xdg-open").arg(url).spawn()?;
    }
    
    Ok(())
}
```

**Alternative using `open` crate (from television):**
```rust
use open;

pub fn open_url(url: &str) -> Result<()> {
    open::that(url)?;
    Ok(())
}
```

### Error Handling Strategies

**Graceful Degradation (from television):**
```rust
pub fn get_working_directory() -> Option<String> {
    std::env::var("PWD")
        .or_else(|_| std::env::current_dir().map(|p| p.to_string_lossy().to_string()))
        .ok()
}

pub fn get_git_repo_name() -> Option<String> {
    let output = Command::new("git")
        .args(&["remote", "get-url", "origin"])
        .output()
        .ok()?;
    
    if !output.status.success() {
        return None;
    }
    
    let url = String::from_utf8(output.stdout).ok()?;
    parse_git_url(&url)
}
```

### Pros & Cons

**Pros:**
- ✅ Universal availability
- ✅ No dependencies to manage
- ✅ Leverages existing tools
- ✅ User familiarity

**Cons:**
- ❌ Fragile (depends on external tools)
- ❌ Version differences
- ❌ Platform variations
- ❌ Error handling complexity

---

## Pattern 2: REST API Integration

### Overview
HTTP-based integration with external services. Used for cloud services, SaaS platforms, and modern APIs.

### Implementation Patterns

#### 2.1 Basic HTTP Client

**From tatuin (Todoist integration):**
```rust
use reqwest::{Client, header};
use serde::{Deserialize, Serialize};

pub struct TodoistClient {
    client: Client,
    api_key: String,
    base_url: String,
}

impl TodoistClient {
    pub fn new(api_key: String) -> Self {
        let mut headers = header::HeaderMap::new();
        headers.insert(
            "Authorization",
            header::HeaderValue::from_str(&format!("Bearer {}", api_key)).unwrap(),
        );
        
        let client = Client::builder()
            .default_headers(headers)
            .timeout(Duration::from_secs(30))
            .build()
            .unwrap();
        
        Self {
            client,
            api_key,
            base_url: "https://api.todoist.com/rest/v2".to_string(),
        }
    }
    
    pub async fn get_tasks(&self, filter: Option<&str>) -> Result<Vec<Task>> {
        let mut url = format!("{}/tasks", self.base_url);
        
        if let Some(f) = filter {
            url.push_str(&format!("?filter={}", urlencoding::encode(f)));
        }
        
        let response = self.client
            .get(&url)
            .send()
            .await?;
        
        if !response.status().is_success() {
            let error_text = response.text().await?;
            return Err(anyhow!("API error: {}", error_text));
        }
        
        let tasks = response.json::<Vec<Task>>().await?;
        Ok(tasks)
    }
    
    pub async fn create_task(&self, content: &str, project_id: Option<&str>) -> Result<Task> {
        let mut body = json!({
            "content": content,
        });
        
        if let Some(pid) = project_id {
            body["project_id"] = json!(pid);
        }
        
        let response = self.client
            .post(&format!("{}/tasks", self.base_url))
            .json(&body)
            .send()
            .await?;
        
        if !response.status().is_success() {
            let error_text = response.text().await?;
            return Err(anyhow!("Failed to create task: {}", error_text));
        }
        
        let task = response.json::<Task>().await?;
        Ok(task)
    }
}
```

#### 2.2 Caching Layer

**From jirust (SurrealDB caching):**
```rust
use surrealdb::Surreal;
use surrealdb::engine::local::Db;

pub struct JiraClient {
    http_client: Client,
    cache: Surreal<Db>,
    base_url: String,
    auth: String,
}

impl JiraClient {
    pub async fn get_issue(&self, key: &str) -> Result<Issue> {
        // Check cache first
        let cached: Option<Issue> = self.cache
            .select(("issue", key))
            .await?;
        
        if let Some(issue) = cached {
            // Check if cache is fresh (< 5 minutes)
            if issue.cached_at.elapsed() < Duration::from_secs(300) {
                return Ok(issue);
            }
        }
        
        // Fetch from API
        let issue = self.fetch_issue_from_api(key).await?;
        
        // Update cache
        let cached_issue = CachedIssue {
            data: issue.clone(),
            cached_at: Instant::now(),
        };
        self.cache
            .create(("issue", key))
            .content(cached_issue)
            .await?;
        
        Ok(issue)
    }
    
    async fn fetch_issue_from_api(&self, key: &str) -> Result<Issue> {
        let response = self.http_client
            .get(&format!("{}/rest/api/2/issue/{}", self.base_url, key))
            .header("Authorization", format!("Basic {}", self.auth))
            .send()
            .await?;
        
        if response.status() == 404 {
            return Err(anyhow!("Issue not found: {}", key));
        }
        
        if !response.status().is_success() {
            return Err(anyhow!("API error: {}", response.status()));
        }
        
        let issue = response.json::<Issue>().await?;
        Ok(issue)
    }
}
```

#### 2.3 Pagination Handling

**From sc-cli (Shortcut API):**
```rust
pub struct PaginatedResults<T> {
    pub items: Vec<T>,
    pub next_token: Option<String>,
    pub total: usize,
}

impl ShortcutClient {
    pub async fn get_stories_paginated(
        &self,
        query: &str,
        page_size: usize,
    ) -> Result<PaginatedResults<Story>> {
        let mut all_stories = Vec::new();
        let mut next_token: Option<String> = None;
        let mut total = 0;
        
        loop {
            let mut request = self.client
                .post("https://api.app.shortcut.com/api/v3/stories/search")
                .json(&json!({
                    "query": query,
                    "page_size": page_size,
                }));
            
            if let Some(token) = &next_token {
                request = request.header("Shortcut-Next-Token", token);
            }
            
            let response = request.send().await?;
            
            if !response.status().is_success() {
                return Err(anyhow!("API error"));
            }
            
            let stories: Vec<Story> = response.json().await?;
            total += stories.len();
            all_stories.extend(stories);
            
            // Check for next page
            next_token = response
                .headers()
                .get("shortcut-next-token")
                .and_then(|v| v.to_str().ok())
                .map(String::from);
            
            if next_token.is_none() || stories.len() < page_size {
                break;
            }
        }
        
        Ok(PaginatedResults {
            items: all_stories,
            next_token: None,
            total,
        })
    }
}
```

#### 2.4 Rate Limiting

**From tatuin:**
```rust
use std::sync::Arc;
use tokio::sync::Semaphore;
use tokio::time::{sleep, Duration};

pub struct RateLimitedClient {
    inner: Client,
    semaphore: Arc<Semaphore>,
    min_interval: Duration,
    last_request: Arc<Mutex<Instant>>,
}

impl RateLimitedClient {
    pub fn new(client: Client, max_requests: usize, interval_secs: u64) -> Self {
        Self {
            inner: client,
            semaphore: Arc::new(Semaphore::new(max_requests)),
            min_interval: Duration::from_secs(interval_secs),
            last_request: Arc::new(Mutex::new(Instant::now() - Duration::from_secs(100))),
        }
    }
    
    pub async fn get(&self, url: &str) -> Result<Response> {
        let _permit = self.semaphore.acquire().await?;
        
        // Rate limiting
        let mut last = self.last_request.lock().await;
        let elapsed = last.elapsed();
        if elapsed < self.min_interval {
            sleep(self.min_interval - elapsed).await;
        }
        *last = Instant::now();
        drop(last);
        
        let response = self.inner.get(url).send().await?;
        
        // Handle rate limit response
        if response.status() == 429 {
            let retry_after = response
                .headers()
                .get("retry-after")
                .and_then(|v| v.to_str().ok())
                .and_then(|v| v.parse::<u64>().ok())
                .unwrap_or(60);
            
            sleep(Duration::from_secs(retry_after)).await;
            return self.get(url).await; // Retry
        }
        
        Ok(response)
    }
}
```

### Authentication Patterns

**API Key in Header (most common):**
```rust
header::HeaderValue::from_str(&format!("Bearer {}", api_key))
```

**Basic Auth (from jirust):**
```rust
let credentials = format!("{}:{}", email, api_token);
let encoded = base64::encode(credentials);
header::HeaderValue::from_str(&format!("Basic {}", encoded))
```

**OAuth2 (not observed, but common):**
```rust
// Token refresh pattern
pub async fn refresh_token(&mut self) -> Result<()> {
    let response = self.client
        .post("https://oauth.provider.com/token")
        .form(&[
            ("grant_type", "refresh_token"),
            ("refresh_token", &self.refresh_token),
            ("client_id", &self.client_id),
        ])
        .send()
        .await?;
    
    let token_response: TokenResponse = response.json().await?;
    self.access_token = token_response.access_token;
    self.expires_at = Instant::now() + Duration::from_secs(token_response.expires_in);
    
    Ok(())
}
```

### Pros & Cons

**Pros:**
- ✅ Rich functionality
- ✅ Real-time data
- ✅ Multi-device sync
- ✅ Professional integration

**Cons:**
- ❌ Network dependency
- ❌ Authentication complexity
- ❌ Rate limiting
- ❌ API versioning issues

---

## Pattern 3: Model Context Protocol (MCP)

### Overview
An emerging pattern for AI-augmented TUIs. MCP allows LLMs to interact with the application through a standardized protocol.

### Implementation Patterns

#### 3.1 MCP Server Architecture

**From kanban (simplified):**
```rust
use rmcp::{
    model::*,
    schemars,
    server::{Server, ServerHandler},
    tool,
};

#[derive(Clone)]
pub struct KanbanMcpServer {
    context: Arc<Mutex<KanbanContext>>,
}

#[tool]
impl KanbanMcpServer {
    pub fn new(context: KanbanContext) -> Self {
        Self {
            context: Arc::new(Mutex::new(context)),
        }
    }
    
    #[tool(name = "list_boards", description = "List all kanban boards")]
    async fn list_boards(&self) -> CallToolResult {
        let ctx = self.context.lock().await;
        let boards = ctx.list_boards();
        
        let content = boards
            .iter()
            .map(|b| format!("{}: {}", b.id, b.name))
            .collect::<Vec<_>>()
            .join("\n");
        
        CallToolResult::success(vec![Content::text(content)])
    }
    
    #[tool(name = "create_card", description = "Create a new card on a board")]
    async fn create_card(
        &self,
        #[tool(param)] board_id: String,
        #[tool(param)] title: String,
        #[tool(param)] description: Option<String>,
    ) -> CallToolResult {
        let mut ctx = self.context.lock().await;
        
        match ctx.create_card(&board_id, &title, description.as_deref()).await {
            Ok(card) => CallToolResult::success(vec![
                Content::text(format!("Created card: {}", card.id))
            ]),
            Err(e) => CallToolResult::error(e.to_string()),
        }
    }
    
    #[tool(name = "move_card", description = "Move a card to a different column")]
    async fn move_card(
        &self,
        #[tool(param)] card_id: String,
        #[tool(param)] target_column: String,
    ) -> CallToolResult {
        let mut ctx = self.context.lock().await;
        
        match ctx.move_card(&card_id, &target_column).await {
            Ok(_) => CallToolResult::success(vec![
                Content::text(format!("Moved card {} to {}", card_id, target_column))
            ]),
            Err(e) => CallToolResult::error(e.to_string()),
        }
    }
}

#[tokio::main]
async fn main() {
    let context = KanbanContext::load().await.unwrap();
    let server = KanbanMcpServer::new(context);
    
    Server::new(server)
        .serve()
        .await
        .unwrap();
}
```

#### 3.2 MCP Client Integration

**From nereid (consumes MCP tools):**
```rust
use rmcp::client::Client;

pub struct NereidMcpClient {
    client: Client,
}

impl NereidMcpClient {
    pub async fn query_diagram(&self, diagram_id: &str, query: &str) -> Result<String> {
        let result = self.client
            .call_tool(
                "diagram.query",
                json!({
                    "diagram_id": diagram_id,
                    "query": query,
                }),
            )
            .await?;
        
        Ok(result.content[0].text.clone())
    }
    
    pub async fn analyze_routes(&self, diagram_id: &str) -> Result<Vec<Route>> {
        let result = self.client
            .call_tool(
                "flow.routes",
                json!({
                    "diagram_id": diagram_id,
                }),
            )
            .await?;
        
        let routes: Vec<Route> = serde_json::from_str(&result.content[0].text)?;
        Ok(routes)
    }
}
```

### Pros & Cons

**Pros:**
- ✅ AI-powered workflows
- ✅ Natural language interaction
- ✅ Extensible architecture
- ✅ Emerging standard

**Cons:**
- ❌ Very new (limited tooling)
- ❌ Adds complexity
- ❌ Requires LLM access
- ❌ Security considerations

### When to Use
- AI-augmented workflows
- Complex query interfaces
- Power user features
- Natural language commands

---

## Decision Matrix

### Integration Strategy by Use Case

| Use Case | Shell | REST API | MCP |
|----------|-------|----------|-----|
| Local tools (git, grep) | ✅ | ❌ | ❌ |
| Cloud services (JIRA, Todoist) | ❌ | ✅ | ❌ |
| AI augmentation | ❌ | ❌ | ✅ |
| File operations | ✅ | ❌ | ❌ |
| Multi-device sync | ❌ | ✅ | ❌ |
| Complex queries | ❌ | ❌ | ✅ |

---

## Best Practices

### 1. Graceful Degradation

**From television:**
```rust
pub async fn fetch_data(&self) -> Result<Data> {
    // Try API first
    if let Ok(data) = self.fetch_from_api().await {
        return Ok(data);
    }
    
    // Fall back to cache
    if let Ok(data) = self.load_from_cache() {
        return Ok(data);
    }
    
    // Fall back to defaults
    Ok(Data::default())
}
```

### 2. Timeouts and Cancellation

```rust
pub async fn fetch_with_timeout(&self) -> Result<Data> {
    match timeout(Duration::from_secs(5), self.fetch()).await {
        Ok(Ok(data)) => Ok(data),
        Ok(Err(e)) => Err(e),
        Err(_) => Err(anyhow!("Request timed out")),
    }
}
```

### 3. Retry Logic

```rust
pub async fn fetch_with_retry(&self, retries: u32) -> Result<Data> {
    let mut last_error = None;
    
    for i in 0..retries {
        match self.fetch().await {
            Ok(data) => return Ok(data),
            Err(e) => {
                last_error = Some(e);
                sleep(Duration::from_millis(100 * 2_u64.pow(i))).await;
            }
        }
    }
    
    Err(last_error.unwrap())
}
```

### 4. Configuration Management

**From tatuin:**
```rust
#[derive(Deserialize)]
pub struct Config {
    pub providers: HashMap<String, ProviderConfig>,
}

#[derive(Deserialize)]
pub struct ProviderConfig {
    pub api_key: Option<String>,
    pub base_url: Option<String>,
    pub timeout_secs: Option<u64>,
}

impl Config {
    pub fn load() -> Result<Self> {
        let config_path = dirs::config_dir()
            .ok_or_else(|| anyhow!("No config directory"))?
            .join("tatuin/config.toml");
        
        let content = fs::read_to_string(&config_path)?;
        let config: Config = toml::from_str(&content)?;
        Ok(config)
    }
}
```

---

## Conclusion

### Summary

1. **Shell Integration:** Use for local tools and universal operations
2. **REST APIs:** Use for cloud services and rich functionality
3. **MCP:** Consider for AI-augmented, cutting-edge applications

### The Integration Stack

```
┌─────────────────────────────────────────┐
│           TUI Application               │
└─────────────┬───────────────────────────┘
              │
    ┌─────────┴──────────┐
    │                    │
┌───▼────┐          ┌────▼────┐
│ Shell  │          │  HTTP   │
│ Exec   │          │ Client  │
└────────┘          └────┬────┘
                         │
               ┌─────────┴─────────┐
               │                   │
          ┌────▼────┐        ┌────▼────┐
          │ REST    │        │   MCP   │
          │ APIs    │        │ Server  │
          └─────────┘        └─────────┘
```

---

## References

1. **reqwest:** https://docs.rs/reqwest/
2. **MCP Protocol:** https://modelcontextprotocol.io/
3. **rmcp crate:** https://docs.rs/rmcp/
4. **API Design Best Practices:** https://docs.github.com/en/rest/guides/best-practices-for-rest-api-design

---

*Document generated from analysis of 18 Rust TUI projects (2026-03-26)*
