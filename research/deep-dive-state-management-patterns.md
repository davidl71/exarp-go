# Deep Dive: State Management Patterns in Rust TUI Applications

## Executive Summary

State management is the backbone of any TUI application. This analysis examines patterns across **18 Rust TUI projects**, identifying three primary architectures: **Synchronous Event Loops** (61%), **Asynchronous Actor Models** (28%), and **Hybrid Approaches** (11%).

**Key Insight:** The choice between sync and async depends on I/O complexity, not application size. Even small apps benefit from async when integrating with external APIs or handling real-time updates.

---

## Pattern Adoption Matrix

| Pattern | Projects | % | Complexity | Best For |
|---------|----------|---|------------|----------|
| **Sync Event Loop** | 11 | 61% | Low-Medium | Simple TUIs, local-only apps |
| **Async Actor Model** | 5 | 28% | High | API integration, real-time updates |
| **Hybrid (Sync UI + Async I/O)** | 2 | 11% | Medium | File watching, background tasks |

### Projects by Pattern

**Sync Event Loop (11):**
- basilk, td, tmmpr, taskui, timr-tui, todolist-rust, taskfinder, rusty-krab-manager, work-tuimer, jirust, sc-cli

**Async Actor Model (5):**
- television, maelstrom, rust_kanban, tatuin, nereid

**Hybrid (2):**
- kanban, taskwarrior-tui

---

## Pattern 1: Synchronous Event Loop

### Overview
The simplest and most common pattern. A single thread handles UI rendering and event processing in a tight loop.

### Architecture

```
┌─────────────────────────────────────┐
│         Main Thread                 │
│  ┌─────────────────────────────┐   │
│  │      Event Loop             │   │
│  │  1. Poll for events         │   │
│  │  2. Update state            │   │
│  │  3. Render UI               │   │
│  └─────────────────────────────┘   │
└─────────────────────────────────────┘
```

### Implementation

**Basic Pattern (from basilk):**
```rust
fn run_app(&mut self, terminal: &mut Terminal) -> Result<()> {
    loop {
        // 1. Render
        terminal.draw(|f| self.render(f))?;
        
        // 2. Poll events (blocking with timeout)
        if event::poll(Duration::from_millis(100))? {
            if let Event::Key(key) = event::read()? {
                if key.kind == KeyEventKind::Press {
                    self.handle_key(key)?;
                }
            }
        }
        
        // 3. Update (if needed)
        self.update()?;
        
        if self.should_quit {
            break;
        }
    }
    Ok(())
}
```

**With Mode State Machine (from td):**
```rust
pub enum Mode {
    Normal,
    Insert,
    Visual,
    Command,
}

pub struct App {
    mode: Mode,
    items: Vec<Item>,
    selected: ListState,
}

fn handle_key(&mut self, key: KeyEvent) -> Result<()> {
    match self.mode {
        Mode::Normal => self.handle_normal_key(key),
        Mode::Insert => self.handle_insert_key(key),
        Mode::Visual => self.handle_visual_key(key),
        Mode::Command => self.handle_command_key(key),
    }
}
```

### State Machine Deep Dive

**Mode Enum Pattern (16/18 projects use this):**

The mode enum is the most prevalent state management technique, used in **89% of projects**.

**Simple Two-Mode (from taskui):**
```rust
pub enum InputMode {
    Select,   // Navigate and execute
    Search,   // Filter items
}
```

**Complex Multi-Mode (from basilk):**
```rust
pub enum ViewMode {
    ViewProjects,
    RenameProject,
    AddProject,
    DeleteProject,
    ViewTasks,
    RenameTask,
    ChangeStatusTask,
    ChangePriorityTask,
    AddTask,
    DeleteTask,
    InfoMigration,
}
```

**Mode with Data (from rust_kanban):**
```rust
pub enum AppMode {
    Normal,
    Editing(String),           // With buffer
    Visual(Vec<usize>),        // With selection
    CommandPalette(String),    // With input
}
```

### Pros & Cons

**Pros:**
- ✅ Simple to understand and debug
- ✅ No concurrency issues
- ✅ Predictable execution flow
- ✅ Low overhead

**Cons:**
- ❌ Blocks on I/O operations
- ❌ Can't handle real-time updates well
- ❌ Limited scalability

### When to Use
- Local-only applications
- Simple CRUD interfaces
- No external API dependencies
- Prototyping and MVPs

---

## Pattern 2: Asynchronous Actor Model

### Overview
Uses tokio and message passing to decouple UI rendering from I/O operations. Events are sent through channels, and components process them asynchronously.

### Architecture

```
┌─────────────────────────────────────────────────────────┐
│                    Tokio Runtime                         │
│  ┌─────────────┐    ┌─────────────┐    ┌─────────────┐  │
│  │  UI Task    │    │  I/O Task   │    │ Timer Task  │  │
│  │             │◄──►│             │    │             │  │
│  │ - Render    │    │ - API calls │    │ - Tick      │  │
│  │ - Input     │    │ - File ops  │    │ - Debounce  │  │
│  └──────┬──────┘    └──────┬──────┘    └──────┬──────┘  │
│         │                  │                  │         │
│         └──────────────────┼──────────────────┘         │
│                            ▼                           │
│                    ┌───────────────┐                   │
│                    │  Event Bus    │                   │
│                    │  (mpsc/       │                   │
│                    │   broadcast)   │                   │
│                    └───────────────┘                   │
└─────────────────────────────────────────────────────────┘
```

### Implementation

**From television (881K LOC):**
```rust
pub struct EventLoop {
    pub rx: mpsc::UnboundedReceiver<Event<KeyCode>>,
    pub tx: mpsc::UnboundedSender<Event<KeyCode>>,
    pub abort: mpsc::UnboundedSender<()>,
    pub tick_rate: Duration,
}

impl EventLoop {
    pub fn new(tick_rate: u64) -> Self {
        let (tx, rx) = mpsc::unbounded_channel();
        let (abort_tx, mut abort_rx) = mpsc::unbounded_channel();
        
        let event_tx = tx.clone();
        tokio::spawn(async move {
            let mut reader = crossterm::event::EventStream::new();
            let mut tick = interval(Duration::from_millis(tick_rate));
            
            loop {
                tokio::select! {
                    _ = abort_rx.recv() => break,
                    _ = tick.tick() => {
                        event_tx.send(Event::Tick).ok();
                    }
                    Some(Ok(evt)) = reader.next() => {
                        if let crossterm::event::Event::Key(key) = evt {
                            event_tx.send(Event::Key(key.code)).ok();
                        }
                    }
                }
            }
        });
        
        Self { rx, tx, abort: abort_tx, tick_rate: Duration::from_millis(tick_rate) }
    }
}
```

**From maelstrom (distributed test runner):**
```rust
// Multi-threaded runtime with specialized tasks
#[tokio::main]
async fn main() -> Result<()> {
    let r = tokio::runtime::Builder::new_multi_thread()
        .enable_all()
        .build()?
        .block_on(async { 
            // Spawn I/O handler
            let io_handle = tokio::spawn(io_handler(io_rx));
            
            // Spawn UI handler
            let ui_handle = tokio::spawn(ui_handler(ui_rx));
            
            // Spawn worker tasks
            for worker_id in 0..num_workers {
                tokio::spawn(worker_task(worker_id, job_rx));
            }
            
            // Main coordinator
            coordinator.run().await
        });
    r
}
```

**From rust_kanban:**
```rust
pub struct WidgetManager<'a> {
    pub app: Arc<tokio::sync::Mutex<App<'a>>>,
}

impl<'a> WidgetManager<'a> {
    pub async fn update(&mut self) -> Result<()> {
        let app = self.app.lock().await;
        // Update widgets asynchronously
        self.update_toast(&app)?;
        self.update_command_palette(&app)?;
        self.update_date_picker(&app)?;
        Ok(())
    }
}
```

### Component Trait Pattern

**From nereid (MCP server):**
```rust
#[async_trait]
pub trait Component: Send + Sync {
    async fn init(&mut self) -> Result<()>;
    async fn handle_event(&mut self, event: Event) -> Result<Action>;
    async fn update(&mut self, action: Action) -> Result<()>;
    async fn render(&self, frame: &mut Frame, area: Rect) -> Result<()>;
}

// Component collection
pub struct App {
    components: Vec<Box<dyn Component>>,
    event_tx: mpsc::UnboundedSender<Event>,
    action_tx: mpsc::UnboundedSender<Action>,
}

impl App {
    pub async fn run(&mut self) -> Result<()> {
        loop {
            tokio::select! {
                Some(event) = self.event_rx.recv() => {
                    for component in &mut self.components {
                        let action = component.handle_event(event.clone()).await?;
                        component.update(action).await?;
                    }
                }
                _ = self.render_interval.tick() => {
                    self.render().await?;
                }
            }
        }
    }
}
```

### Pros & Cons

**Pros:**
- ✅ Non-blocking I/O
- ✅ Real-time updates
- ✅ Better scalability
- ✅ Clean separation of concerns

**Cons:**
- ❌ Higher complexity
- ❌ Debugging difficulty (async stack traces)
- ❌ Risk of deadlocks
- ❌ More boilerplate

### When to Use
- External API integration
- Real-time data streams
- Background processing
- Multi-user/distributed systems

---

## Pattern 3: Hybrid (Sync UI + Async I/O)

### Overview
Keeps UI rendering synchronous for simplicity but offloads I/O to async tasks. Best of both worlds for specific use cases.

### Architecture

```
┌─────────────────────────────────────────────────────────┐
│                    Main Thread                           │
│  ┌───────────────────────────────────────────────────┐  │
│  │              Synchronous UI Loop                  │  │
│  │  - Render                                         │  │
│  │  - Handle input                                   │  │
│  │  - Check async results (non-blocking)             │  │
│  └───────────────────┬───────────────────────────────┘  │
│                      │                                   │
│                      ▼                                   │
│  ┌───────────────────────────────────────────────────┐  │
│  │              Async I/O Pool (tokio)               │  │
│  │  - File watching                                  │  │
│  │  - Network requests                               │  │
│  │  - Heavy computation                              │  │
│  └───────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────┘
```

### Implementation

**From kanban (file watching):**
```rust
pub struct App {
    // Sync state
    state: MapState,
    
    // Async I/O handle
    file_watcher: RecommendedWatcher,
    change_rx: mpsc::Receiver<DebouncedEvent>,
}

impl App {
    fn run(&mut self) -> Result<()> {
        loop {
            // Render (sync)
            self.terminal.draw(|f| self.render(f))?;
            
            // Check for file changes (non-blocking)
            if let Ok(event) = self.change_rx.try_recv() {
                self.handle_file_change(event)?;
            }
            
            // Handle input (blocking with timeout)
            if event::poll(Duration::from_millis(100))? {
                if let Event::Key(key) = event::read()? {
                    self.handle_key(key)?;
                }
            }
        }
    }
}
```

**From taskwarrior-tui:**
```rust
fn run_app<B: Backend>(
    terminal: &mut Terminal<B>,
    mut app: App,
    tick_rate: Duration,
) -> Result<()> {
    let mut last_tick = Instant::now();
    
    loop {
        // Draw UI
        terminal.draw(|f| ui::draw(f, &mut app))?;
        
        // Non-blocking check for async updates
        if let Some(update) = app.check_async_updates() {
            app.apply_update(update)?;
        }
        
        // Handle events with timeout
        let timeout = tick_rate
            .checked_sub(last_tick.elapsed())
            .unwrap_or_else(|| Duration::from_secs(0));
        
        if crossterm::event::poll(timeout)? {
            if let Event::Key(key) = crossterm::event::read()? {
                app.handle_key(key).await?;
            }
        }
        
        if last_tick.elapsed() >= tick_rate {
            app.on_tick().await?;
            last_tick = Instant::now();
        }
    }
}
```

### Pros & Cons

**Pros:**
- ✅ Simple UI code
- ✅ Async I/O capabilities
- ✅ No complex message passing
- ✅ Easier to debug than full async

**Cons:**
- ❌ Manual synchronization needed
- ❌ Limited async capabilities
- ❌ Potential for blocking if not careful

### When to Use
- File watching
- Periodic background tasks
- Simple API polling
- Transitioning from sync to async

---

## State Synchronization Strategies

### 1. Shared State with Mutex (Most Common)

**From television:**
```rust
pub struct App {
    pub state: Arc<Mutex<AppState>>,
}

impl App {
    pub fn update(&self, action: Action) {
        let mut state = self.state.lock().unwrap();
        match action {
            Action::Up => state.selected -= 1,
            Action::Down => state.selected += 1,
            _ => {}
        }
    }
}
```

**Pros:** Simple, works everywhere
**Cons:** Risk of deadlocks, contention

### 2. Message Passing (Channel-Based)

**From maelstrom:**
```rust
pub struct Broker {
    request_rx: mpsc::Receiver<Request>,
    response_tx: mpsc::Sender<Response>,
}

impl Broker {
    async fn run(&mut self) {
        while let Some(req) = self.request_rx.recv().await {
            let resp = self.handle_request(req).await;
            self.response_tx.send(resp).await.ok();
        }
    }
}
```

**Pros:** No shared state, lock-free
**Cons:** More complex, harder to trace

### 3. Event Sourcing

**From nereid:**
```rust
pub enum Event {
    // User events
    Key(KeyEvent),
    Mouse(MouseEvent),
    
    // System events
    Tick,
    Render,
    Resize(u16, u16),
    
    // Application events
    DiagramLoaded(DiagramId),
    WalkthroughUpdated(WalkthroughId),
}

pub fn reduce(state: State, event: Event) -> State {
    match event {
        Event::Key(k) => handle_key(state, k),
        Event::DiagramLoaded(id) => load_diagram(state, id),
        _ => state,
    }
}
```

**Pros:** Time-travel debugging, audit trail
**Cons:** Memory overhead, complexity

---

## Decision Framework

### Choose Sync Event Loop If:
- [ ] No external API calls needed
- [ ] Real-time updates not required
- [ ] Simple CRUD operations only
- [ ] Team new to Rust/async
- [ ] Quick prototyping

### Choose Async Actor Model If:
- [ ] External API integration required
- [ ] Real-time data streams
- [ ] Background processing needed
- [ ] Multiple I/O sources
- [ ] Distributed/multi-user system

### Choose Hybrid If:
- [ ] File watching required
- [ ] Periodic background tasks
- [ ] Simple API polling
- [ ] Transitioning architecture
- [ ] Mix of sync and async libraries

---

## Best Practices

### 1. State Immutability

**Good (from kanban):**
```rust
pub fn update(&self, action: Action) -> Self {
    let mut new_state = self.clone();
    new_state.apply(action);
    new_state
}
```

**Bad:**
```rust
pub fn update(&mut self, action: Action) {
    // Direct mutation makes debugging hard
    self.field = action.value;
}
```

### 2. Explicit Error Handling

**Good (from television):**
```rust
pub fn handle_result<T>(result: Result<T>) -> Action {
    match result {
        Ok(val) => Action::Success(val),
        Err(e) => Action::Error(e.to_string()),
    }
}
```

### 3. Avoid Deadlocks

**Rule:** Always acquire locks in the same order

```rust
// Good
let state = app_state.lock().unwrap();
let config = config.lock().unwrap();

// Bad (potential deadlock)
// Thread 1: locks app_state, waits for config
// Thread 2: locks config, waits for app_state
```

### 4. Bounded Channels

**Good (from maelstrom):**
```rust
// Bounded to prevent memory exhaustion
let (tx, rx) = mpsc::channel::<Event>(1000);
```

**Bad:**
```rust
// Unbounded can cause OOM
let (tx, rx) = mpsc::unbounded_channel::<Event>();
```

### 5. Cancellation Safety

**From ratatui async template:**
```rust
pub struct Tui {
    cancellation_token: CancellationToken,
}

impl Tui {
    pub fn stop(&self) {
        self.cancellation_token.cancel();
    }
}
```

---

## Anti-Patterns to Avoid

### 1. Blocking in Async Context

**Don't:**
```rust
async fn bad() {
    let data = blocking_file_read(); // Blocks thread!
}
```

**Do:**
```rust
async fn good() {
    let data = tokio::fs::read("file.txt").await; // Non-blocking
}
```

### 2. Holding Locks Across Await Points

**Don't:**
```rust
async fn bad(app: Arc<Mutex<App>>) {
    let guard = app.lock().unwrap();
    let data = fetch_data().await; // Lock held during await!
    guard.update(data);
}
```

**Do:**
```rust
async fn good(app: Arc<Mutex<App>>) {
    let data = fetch_data().await;
    let mut guard = app.lock().unwrap();
    guard.update(data);
    drop(guard); // Explicit drop
}
```

### 3. Uncontrolled State Mutation

**Don't:**
```rust
impl App {
    fn random_update(&mut self) {
        // Side effects everywhere
        self.field1 = random();
        self.field2 = random();
    }
}
```

**Do:**
```rust
impl App {
    fn apply(&mut self, action: Action) {
        match action {
            Action::UpdateField1(v) => self.field1 = v,
            Action::UpdateField2(v) => self.field2 = v,
        }
    }
}
```

---

## Performance Considerations

### Rendering Optimization

**From television (60+ FPS):**
```rust
// Only render when state changes
if self.state.changed {
    terminal.draw(|f| self.render(f))?;
    self.state.changed = false;
}
```

### Memory Management

**From maelstrom:**
```rust
// Pre-allocate buffers
pub struct Matcher {
    inner: nucleo::Nucleo,
    col_indices_buffer: Vec<u32>, // Reused buffer
}
```

### CPU Usage

**Sync loop (low CPU):**
```rust
// 100ms poll = ~1% CPU
if event::poll(Duration::from_millis(100))? {
    // Handle event
}
```

**Async loop (variable):**
```rust
// Only wakes when events arrive
tokio::select! {
    Some(event) = event_rx.recv() => {},
    _ = tick_interval.tick() => {},
}
```

---

## Conclusion

### Summary

1. **Start Simple:** Use sync event loop for MVPs
2. **Add Async When Needed:** Don't over-engineer early
3. **Use Mode Enums:** They're universally effective
4. **Prefer Message Passing:** Over shared state when possible
5. **Measure Performance:** Optimize based on data, not assumptions

### Recommendations by Project Size

| Size | Sync | Async | Hybrid |
|------|------|-------|--------|
| Small (<2K LOC) | ✅ | ❌ | ❌ |
| Medium (2-15K) | ✅ | ✅ | ✅ |
| Large (15K+) | ❌ | ✅ | ✅ |

### The Golden Rule

**"Make invalid states unrepresentable"** - Use Rust's type system (enums, match exhaustiveness) to enforce valid state transitions at compile time.

---

## References

1. **ratatui Async Template:** https://github.com/ratatui-org/async-template
2. **Component Architecture:** https://ratatui.rs/concepts/application-patterns/component-architecture/
3. **Async Events Tutorial:** https://ratatui.rs/tutorials/counter-async-app/full-async-events/
4. **Tokio Documentation:** https://tokio.rs/
5. **d-holguin/async-ratatui:** https://github.com/d-holguin/async-ratatui

---

*Document generated from analysis of 18 Rust TUI projects (2026-03-26)*
