# MCP Framework Comparison: Go, Rust, C, C++, and Assembly

## Executive Summary

This document compares Model Context Protocol (MCP) frameworks across Go, Rust, C, C++, and Assembly. MCP is an open standard introduced by Anthropic in November 2024 for integrating LLMs with external tools and data sources.

**Key Findings:**
- **Go**: Multiple mature frameworks with good performance and simplicity
- **Rust**: Strong safety guarantees with high performance, enterprise-ready options
- **C/C++**: Enterprise-grade solutions with cross-language support
- **Assembly**: No direct implementations; WebAssembly provides the closest alternative

---

## Go Frameworks

### 1. Foxy Contexts
**Type**: Declarative Go library  
**Focus**: Simplicity and ease of use

**Strengths:**
- Declarative API design
- Simple to learn and use
- Good for rapid prototyping

**Use Cases:**
- Quick MCP server development
- Projects prioritizing developer experience
- Small to medium-scale applications

**Limitations:**
- May lack advanced features compared to other frameworks
- Less suitable for high-throughput scenarios

---

### 2. mcp-go
**Type**: Golang SDK  
**Focus**: Performance and scalability

**Strengths:**
- Supports both MCP servers and clients
- Balanced performance and developer experience
- Good scalability characteristics

**Use Cases:**
- Production MCP servers requiring performance
- Applications needing both client and server capabilities
- Medium to large-scale deployments

**Limitations:**
- Less declarative than Foxy Contexts
- May require more boilerplate code

---

### 3. mcp-golang
**Type**: Type-safe Go framework  
**Focus**: High performance

**Strengths:**
- Type-safe implementation
- Optimized for high-throughput applications
- Strong performance characteristics

**Use Cases:**
- Performance-critical MCP servers
- High-throughput applications
- Systems requiring maximum efficiency

**Limitations:**
- May have steeper learning curve
- Potentially more complex API

---

## Go Frameworks Activity Comparison

### Activity Metrics (as of December 2025)

| Framework | Repository | Stars | Forks | Last Update | Activity Level | Official Status |
| --------- | --------- | ----- | ----- | ----------- | -------------- | --------------- |
| **go-sdk** | modelcontextprotocol/go-sdk | 30+ | N/A | 4 days ago | ⭐⭐⭐⭐⭐ | ✅ Official |
| **mcp-go** | mark3labs/mcp-go | N/A | 730 | 4 weeks ago | ⭐⭐⭐⭐⭐ | ❌ Community |
| **go-mcp** | ThinkInAIXYZ/go-mcp | 104 | 104 | Last month | ⭐⭐⭐⭐ | ❌ Community |
| **mcp-golang** | metoro-io/mcp-golang | N/A | 115 | Active | ⭐⭐⭐ | ❌ Community |
| **Foxy Contexts** | Various | Unknown | Unknown | Unknown | ⭐⭐ | ❌ Community |

### Detailed Activity Analysis

#### 1. go-sdk (Official) - **MOST ACTIVE** ⭐⭐⭐⭐⭐
**Repository**: <https://github.com/modelcontextprotocol/go-sdk>

**Activity Highlights:**
- ✅ **Official SDK** maintained in collaboration with Google
- ✅ **Most recent commits**: 4 days ago (highest activity)
- ✅ **Official support**: Guaranteed alignment with MCP specification
- ✅ **Best for**: Production systems requiring official support

**Why Choose:**
- Official maintenance ensures long-term viability
- Guaranteed compliance with latest MCP specifications
- Best documentation and support
- Recommended for enterprise/production use

**Trade-offs:**
- May be newer/less mature than community alternatives
- Potentially smaller community than mcp-go

---

#### 2. mcp-go (mark3labs) - **HIGHLY ACTIVE** ⭐⭐⭐⭐⭐
**Repository**: <https://github.com/mark3labs/mcp-go>

**Activity Highlights:**
- ✅ **730 forks** - largest community engagement
- ✅ **Last updated**: 4 weeks ago
- ✅ **High community adoption** - most popular community framework
- ✅ **Active development** - core features operational, advanced features in progress

**Why Choose:**
- Largest community and ecosystem
- Extensive real-world usage
- Good for learning from examples
- Strong community support

**Trade-offs:**
- Not official (may lag behind spec updates)
- Community-driven maintenance

---

#### 3. go-mcp (ThinkInAIXYZ) - **ACTIVE** ⭐⭐⭐⭐
**Repository**: <https://github.com/ThinkInAIXYZ/go-mcp>

**Activity Highlights:**
- ✅ **104 stars and forks** - growing community
- ✅ **Last updated**: Last month
- ✅ **Three-layer architecture** - well-designed
- ✅ **Type-safe** with Go's strong type system

**Why Choose:**
- Elegant architecture design
- Good type safety
- Active development
- Growing community

**Trade-offs:**
- Smaller community than mcp-go
- Less adoption than mcp-go

---

#### 4. mcp-golang (metoro-io) - **MODERATE ACTIVITY** ⭐⭐⭐
**Repository**: <https://github.com/metoro-io/mcp-golang>

**Activity Highlights:**
- ✅ **115 forks** - moderate community
- ✅ **Minimal boilerplate** - developer-friendly
- ✅ **Type safety** through native Go structs
- ✅ **Custom transports** - stdio and HTTP support

**Why Choose:**
- Low boilerplate code
- Good for rapid development
- Type-safe design

**Trade-offs:**
- Moderate activity level
- Smaller community
- Less documentation/examples

---

#### 5. Foxy Contexts - **LOW/MODERATE ACTIVITY** ⭐⭐
**Status**: Activity unclear, may be TypeScript-focused

**Activity Highlights:**
- ⚠️ **Limited activity data** available
- ⚠️ **May be primarily TypeScript** - Go support unclear
- ⚠️ **Less community engagement** visible

**Why Choose:**
- Declarative API (if actively maintained)
- Simple to use

**Trade-offs:**
- Unclear maintenance status
- Limited community visibility
- May not be actively developed for Go

---

### Activity-Based Recommendations

#### For Maximum Activity & Official Support:
**Choose: go-sdk (Official)**
- Official maintenance
- Most recent updates
- Guaranteed spec compliance

#### For Largest Community & Ecosystem:
**Choose: mcp-go (mark3labs)**
- 730 forks = largest community
- Most examples and real-world usage
- Strong community support

#### For Balanced Activity & Architecture:
**Choose: go-mcp (ThinkInAIXYZ)**
- Good activity level
- Well-designed architecture
- Growing community

#### For Rapid Development:
**Choose: mcp-golang (metoro-io)**
- Low boilerplate
- Type-safe
- Moderate but active

---

### Activity Trends Summary

1. **Most Active Overall**: `go-sdk` (official) - 4 days ago
2. **Most Community Engagement**: `mcp-go` - 730 forks
3. **Best Architecture**: `go-mcp` - three-layer design
4. **Most Developer-Friendly**: `mcp-golang` - minimal boilerplate
5. **Least Clear**: `Foxy Contexts` - activity uncertain

**Recommendation**: For new projects, choose **go-sdk** (official) for official support, or **mcp-go** (mark3labs) for maximum community and ecosystem support.

---

## Go Frameworks Protocol Support

### MCP Protocol Overview

The Model Context Protocol supports multiple transport mechanisms and protocol versions:

**Transport Protocols:**
- **STDIO**: Standard input/output streams for local process communication
- **HTTP**: Hypertext Transfer Protocol for web-based interactions
- **SSE (Server-Sent Events)**: Real-time, server-initiated updates over HTTP
- **HTTP POST**: Traditional request-response over HTTP

**Protocol Versions:**
- **2025-06-18**: Latest specification (most features)
- **2025-03-26**: Previous stable version
- **Automatic negotiation**: Some frameworks support version negotiation

### Protocol Support Matrix

| Framework | STDIO | HTTP | SSE | HTTP POST | Protocol Version | Version Negotiation |
| --------- | ----- | ---- | --- | --------- | ---------------- | ------------------- |
| **go-sdk** (Official) | ✅ | ✅ | ✅ | ✅ | 2025-06-18+ | ✅ |
| **mcp-go** (mark3labs) | ✅ | ✅ | ✅ | ✅ | 2025-06-18+ | ✅ |
| **go-mcp** (ThinkInAIXYZ) | ✅ | ✅ | ✅ | ✅ | 2025-06-18+ | ✅ |
| **mcp-golang** (metoro-io) | ✅ | ✅ | ⚠️ | ✅ | 2025-06-18+ | ⚠️ |
| **Foxy Contexts** | ✅ | ⚠️ | ⚠️ | ⚠️ | Unknown | ❌ |

**Legend:**
- ✅ = Fully Supported
- ⚠️ = Partial/Unclear Support
- ❌ = Not Supported

### Detailed Protocol Support by Framework

#### 1. go-sdk (Official)
**Repository**: <https://github.com/modelcontextprotocol/go-sdk>

**Transport Protocols:**
- ✅ **STDIO**: Full support for standard input/output communication
- ✅ **HTTP**: Complete HTTP transport implementation
- ✅ **SSE**: Server-Sent Events for real-time updates
- ✅ **HTTP POST**: Traditional HTTP request-response

**Protocol Versions:**
- ✅ **2025-06-18**: Latest specification fully supported
- ✅ **Version Negotiation**: Automatic protocol version negotiation
- ✅ **Backward Compatibility**: Supports older protocol versions

**Features:**
- Official implementation ensures spec compliance
- All transport methods officially supported
- Best protocol version support

**Use Cases:**
- Production systems requiring full protocol support
- Applications needing multiple transport options
- Systems requiring latest MCP features

---

#### 2. mcp-go (mark3labs)
**Repository**: <https://github.com/mark3labs/mcp-go>

**Transport Protocols:**
- ✅ **STDIO**: Full support, primary transport method
- ✅ **HTTP**: Complete HTTP implementation
- ✅ **SSE**: Server-Sent Events supported
- ✅ **HTTP POST**: Standard HTTP POST requests

**Protocol Versions:**
- ✅ **2025-06-18**: Latest specification supported
- ✅ **Version Negotiation**: Automatic negotiation supported
- ✅ **Dynamic Registration**: Advanced capability registration

**Features:**
- Multiple transport methods
- Session management
- Advanced composition patterns
- Dynamic capability filtering

**Use Cases:**
- Applications requiring multiple transport options
- Systems needing dynamic capability management
- High-performance MCP servers

---

#### 3. go-mcp (ThinkInAIXYZ)
**Repository**: <https://github.com/ThinkInAIXYZ/go-mcp>

**Transport Protocols:**
- ✅ **STDIO**: Supported via three-layer architecture
- ✅ **HTTP**: Full HTTP support
- ✅ **SSE**: Server-Sent Events via HTTP layer
- ✅ **HTTP POST**: Standard POST requests

**Protocol Versions:**
- ✅ **2025-06-18**: Latest specification
- ✅ **Version Negotiation**: Supported
- ✅ **JSON-RPC 2.0**: Unified response handling

**Architecture:**
- **Transport Layer**: Manages underlying communication protocols
- **Protocol Layer**: Implements MCP protocol functionality
- **User Layer**: Provides server/client APIs

**Features:**
- Three-layer architecture for modularity
- Type-safe API interfaces
- Resource subscription capabilities
- Pagination support for list APIs

**Use Cases:**
- Applications requiring modular architecture
- Systems needing type-safe interfaces
- Projects requiring resource subscriptions

---

#### 4. mcp-golang (metoro-io)
**Repository**: <https://github.com/metoro-io/mcp-golang>

**Transport Protocols:**
- ✅ **STDIO**: Primary transport, fully supported
- ✅ **HTTP**: Custom HTTP transport support
- ⚠️ **SSE**: Support unclear, may be limited
- ✅ **HTTP POST**: Standard POST requests

**Protocol Versions:**
- ✅ **2025-06-18**: Latest specification supported
- ⚠️ **Version Negotiation**: May have limited support

**Features:**
- Custom transport support (stdio and HTTP)
- Type safety through native Go structs
- Automatic MCP endpoint generation
- Minimal boilerplate code

**Use Cases:**
- Rapid development with stdio transport
- Applications primarily using stdio
- Projects requiring minimal setup

**Limitations:**
- SSE support may be limited
- Version negotiation unclear

---

#### 5. Foxy Contexts
**Status**: Protocol support unclear

**Transport Protocols:**
- ✅ **STDIO**: Likely supported (standard for MCP)
- ⚠️ **HTTP**: Support unclear
- ⚠️ **SSE**: Support unclear
- ⚠️ **HTTP POST**: Support unclear

**Protocol Versions:**
- ❓ **Unknown**: Protocol version support not documented

**Features:**
- Declarative API design
- Simplified server creation

**Limitations:**
- Limited documentation on protocol support
- Transport method support unclear
- Not recommended for production without verification

---

### Protocol Selection Guidelines

#### Choose STDIO if:
- ✅ Local process communication
- ✅ Simple deployment (single binary)
- ✅ No network requirements
- ✅ CLI tool integration

**Best Frameworks**: All Go frameworks support STDIO well, especially:
- **mcp-golang**: Optimized for stdio
- **go-sdk**: Official stdio support
- **mcp-go**: Excellent stdio implementation

---

#### Choose HTTP/SSE if:
- ✅ Network-based communication
- ✅ Real-time updates needed
- ✅ Web application integration
- ✅ Remote server deployment

**Best Frameworks**:
- **go-sdk**: Official HTTP/SSE support
- **mcp-go**: Full HTTP/SSE implementation
- **go-mcp**: Three-layer architecture with HTTP/SSE

---

#### Choose HTTP POST if:
- ✅ Traditional request-response patterns
- ✅ REST-like API integration
- ✅ Simple HTTP communication
- ✅ No real-time requirements

**Best Frameworks**:
- **go-sdk**: Official HTTP POST support
- **mcp-go**: Complete HTTP implementation
- **go-mcp**: HTTP POST via transport layer

---

### Protocol Version Compatibility

**Latest Protocol (2025-06-18) Support:**
- ✅ **go-sdk**: Official, guaranteed support
- ✅ **mcp-go**: Full support with latest features
- ✅ **go-mcp**: Complete implementation
- ✅ **mcp-golang**: Supported
- ❓ **Foxy Contexts**: Unknown

**Version Negotiation:**
- ✅ **go-sdk**: Automatic negotiation
- ✅ **mcp-go**: Automatic negotiation
- ✅ **go-mcp**: Supported
- ⚠️ **mcp-golang**: May have limitations
- ❌ **Foxy Contexts**: Not supported

---

### Recommendations by Protocol Needs

**For Full Protocol Support:**
1. **go-sdk** (Official) - Best choice for complete protocol support
2. **mcp-go** (mark3labs) - Excellent community implementation
3. **go-mcp** (ThinkInAIXYZ) - Well-architected with full support

**For STDIO-Only Projects:**
1. **mcp-golang** - Optimized for stdio
2. **go-sdk** - Official stdio support
3. **mcp-go** - Excellent stdio implementation

**For HTTP/SSE Requirements:**
1. **go-sdk** - Official HTTP/SSE support
2. **mcp-go** - Full HTTP/SSE features
3. **go-mcp** - Modular HTTP/SSE architecture

**For Latest Protocol Features:**
1. **go-sdk** - Official, always up-to-date
2. **mcp-go** - Active development, latest features
3. **go-mcp** - Modern architecture, latest spec

---

## Rust Frameworks

### 1. Turul MCP Framework
**Type**: Comprehensive Rust framework  
**Focus**: Modern patterns, enterprise features

**Strengths:**
- Fully compliant with MCP 2025-06-18 specification
- Serverless deployment support
- Comprehensive testing suites
- Extensive tooling
- Modern design patterns
- Memory safety without garbage collection

**Use Cases:**
- Enterprise-grade MCP servers
- Serverless deployments
- Security-critical applications
- Production systems requiring reliability

**Limitations:**
- Rust learning curve
- May be overkill for simple use cases

**GitHub**: <https://github.com/aussierobots/turul-mcp-framework>

---

### 2. mcp-rs-template
**Type**: CLI server template  
**Focus**: Foundation for MCP server development

**Strengths:**
- Provides solid starting point
- Leverages Rust's safety guarantees
- Good performance characteristics
- Security-focused

**Use Cases:**
- Starting new MCP server projects
- Projects requiring Rust's safety features
- CLI-based MCP servers

**Limitations:**
- Template-based, requires customization
- Less feature-complete than Turul

---

### 3. sakura-mcp
**Type**: Enterprise-grade implementation (Rust + Scala)  
**Focus**: Robust and scalable MCP services

**Strengths:**
- Enterprise-grade reliability
- Combines Rust and Scala strengths
- High performance
- Scalable architecture

**Use Cases:**
- Large-scale enterprise deployments
- Systems requiring maximum reliability
- Complex multi-language integrations

**Limitations:**
- Requires knowledge of both Rust and Scala
- More complex setup and deployment

---

## C/C++ Frameworks

### 1. mcp-cpp
**Type**: MCP server for C/C++ codebases  
**Focus**: Large codebase analysis

**Strengths:**
- Tailored for C/C++ codebases
- Utilizes `clangd` for semantic analysis
- Supports CMake and Meson build systems
- Integration with Claude CLI and Amazon Q Developer CLI
- Excellent for code analysis

**Use Cases:**
- Large C/C++ codebase analysis
- Integration with existing C/C++ projects
- Semantic code analysis
- IDE integration

**Limitations:**
- C/C++ specific, less general-purpose
- Requires clangd infrastructure

**GitHub**: <https://github.com/mpsm/mcp-cpp>

---

### 2. Gopher MCP
**Type**: C++ SDK with C API  
**Focus**: Enterprise security and cross-language support

**Strengths:**
- Enterprise-grade security
- Visibility and connectivity features
- Stable C API for cross-language integration
- Thread-safe dispatcher model
- Modular filter chain architecture
- Supports Python, Go, Rust, Java bindings

**Use Cases:**
- Enterprise security-critical applications
- Multi-language integrations
- Systems requiring C API compatibility
- High-performance production systems

**Limitations:**
- C++ complexity
- Requires C API knowledge for bindings

**GitHub**: <https://github.com/GopherSecurity/gopher-mcp>

---

## Assembly

### Direct Assembly Implementations
**Status**: No widely recognized MCP frameworks exist in pure assembly

**Reasons:**
- Extremely low-level nature
- High complexity of MCP protocol
- Development time and maintenance burden
- Limited practical benefits over compiled languages

### WebAssembly Alternatives

Since direct assembly implementations don't exist, WebAssembly (Wasm) provides the closest alternative for low-level, portable MCP components:

#### 1. Wassette
**Type**: WebAssembly runtime for AI tools  
**Focus**: Secure, sandboxed execution

**Strengths:**
- Secure sandboxed environment
- Fine-grained permission controls
- Supports multiple languages (Rust, JavaScript, Python, Go)
- Portable across platforms
- Security-focused design

**Use Cases:**
- Untrusted MCP tool execution
- Security-critical environments
- Cross-platform deployments
- Multi-language tool composition

**Limitations:**
- WebAssembly overhead
- Requires Wasm compilation

**Reference**: <https://rye.dev/blog/wassette-webassembly-mcp-runtime/>

---

#### 2. Wasmcp
**Type**: Polyglot framework for WebAssembly components  
**Focus**: Composable MCP servers

**Strengths:**
- In-memory composition of components
- Standalone MCP server deployment
- Cross-language component integration
- Modular architecture
- Runtime flexibility

**Use Cases:**
- Composing MCP servers from multiple languages
- Modular tool development
- Standalone deployments
- Language-agnostic tool development

**Limitations:**
- WebAssembly compilation step
- Runtime dependency

**Reference**: <https://spinframework.dev/blog/mcp-with-wasmcp>

---

#### 3. MCP-SandboxScan
**Type**: WebAssembly/WASI sandbox for MCP tools  
**Focus**: Security auditing

**Strengths:**
- Safe execution of untrusted MCP tools
- Auditable reports of external-to-sink exposures
- Addresses security concerns in tool-augmented LLM agents
- WASI compatibility

**Use Cases:**
- Security auditing of MCP tools
- Untrusted tool execution
- Compliance and security reporting

**Limitations:**
- Specialized for security use cases
- May have performance overhead

**Reference**: <https://arxiv.org/abs/2601.01241>

---

## Comparison Matrix

| Framework | Language | Performance | Safety | Ease of Use | Enterprise Features | Cross-Lang Support | Activity Level |
| --------- | -------- | ----------- | ------ | ----------- | ------------------- | ------------------ | -------------- |
| **go-sdk** (Official) | Go | ⭐⭐⭐⭐ | ⭐⭐⭐ | ⭐⭐⭐⭐ | ⭐⭐⭐⭐ | ⭐⭐⭐ | ⭐⭐⭐⭐⭐ |
| mcp-go | Go | ⭐⭐⭐⭐ | ⭐⭐⭐ | ⭐⭐⭐⭐ | ⭐⭐⭐ | ⭐⭐⭐ | ⭐⭐⭐⭐⭐ |
| go-mcp | Go | ⭐⭐⭐⭐ | ⭐⭐⭐ | ⭐⭐⭐⭐ | ⭐⭐⭐ | ⭐⭐⭐ | ⭐⭐⭐⭐ |
| mcp-golang | Go | ⭐⭐⭐⭐⭐ | ⭐⭐⭐ | ⭐⭐⭐ | ⭐⭐⭐ | ⭐⭐ | ⭐⭐⭐ |
| Foxy Contexts | Go | ⭐⭐⭐ | ⭐⭐⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐ | ⭐⭐ | ⭐⭐ |
| Turul MCP | Rust | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐⭐ | ⭐⭐⭐⭐ |
| mcp-rs-template | Rust | ⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐⭐ | ⭐⭐ | ⭐⭐ | ⭐⭐⭐ |
| sakura-mcp | Rust+Scala | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐⭐ | ⭐⭐⭐ |
| mcp-cpp | C++ | ⭐⭐⭐⭐⭐ | ⭐⭐ | ⭐⭐ | ⭐⭐⭐ | ⭐⭐ | ⭐⭐⭐ |
| Gopher MCP | C++ | ⭐⭐⭐⭐⭐ | ⭐⭐⭐ | ⭐⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐⭐ |
| Wassette | Wasm | ⭐⭐⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐⭐ | ⭐⭐⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐⭐ |
| Wasmcp | Wasm | ⭐⭐⭐ | ⭐⭐⭐⭐ | ⭐⭐⭐ | ⭐⭐⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐⭐ |

**Legend**: ⭐ = Low, ⭐⭐⭐ = Medium, ⭐⭐⭐⭐⭐ = High

---

## Selection Guidelines

### Choose Go if:
- ✅ You need rapid development
- ✅ Your team is familiar with Go
- ✅ You want good performance with simplicity
- ✅ Single binary deployment is important
- ✅ You're building medium-scale applications

**Recommended by Activity**:
- **go-sdk (Official)**: For official support and latest spec compliance (most active, 4 days ago)
- **mcp-go (mark3labs)**: For largest community and ecosystem (730 forks, highly active)
- **go-mcp (ThinkInAIXYZ)**: For well-designed architecture (104 stars, active)
- **mcp-golang (metoro-io)**: For minimal boilerplate (115 forks, moderate activity)
- **Foxy Contexts**: Not recommended due to unclear activity status

---

### Choose Rust if:
- ✅ Security and memory safety are critical
- ✅ You need maximum performance
- ✅ You're building enterprise-grade systems
- ✅ Serverless deployment is required
- ✅ You want modern, well-tested frameworks

**Recommended**: Use **Turul MCP Framework** for comprehensive features, or **mcp-rs-template** for starting projects.

---

### Choose C/C++ if:
- ✅ You're integrating with existing C/C++ codebases
- ✅ You need maximum performance
- ✅ Cross-language C API is required
- ✅ Enterprise security features are needed
- ✅ You're building system-level integrations

**Recommended**: Use **Gopher MCP** for enterprise features, or **mcp-cpp** for C/C++ codebase analysis.

---

### Choose WebAssembly if:
- ✅ Security and sandboxing are priorities
- ✅ You need cross-language tool composition
- ✅ Portability across platforms is important
- ✅ You want to write tools in multiple languages
- ✅ Untrusted code execution is required

**Recommended**: Use **Wassette** for security-focused execution, or **Wasmcp** for composable servers.

---

## Performance Considerations

### Raw Performance (Highest to Lowest)
1. **C/C++ frameworks** (mcp-cpp, Gopher MCP) - Native compilation, zero overhead
2. **Rust frameworks** (Turul, sakura-mcp) - Near-native performance with safety
3. **Go frameworks** (mcp-golang, mcp-go) - Good performance with GC overhead
4. **WebAssembly** (Wassette, Wasmcp) - Portable but with Wasm overhead

### Development Speed (Fastest to Slowest)
1. **Go frameworks** - Fast compilation, simple syntax
2. **Rust frameworks** - Good tooling but steeper learning curve
3. **C/C++ frameworks** - Longer development cycles
4. **WebAssembly** - Additional compilation step

### Safety (Highest to Lowest)
1. **Rust frameworks** - Compile-time memory safety
2. **WebAssembly** - Runtime sandboxing
3. **Go frameworks** - GC-managed memory safety
4. **C/C++ frameworks** - Manual memory management

---

## Migration Considerations

### From Python (Current Project)
- **Go**: Easiest migration path, similar development experience
- **Rust**: Requires learning new language, but strong safety benefits
- **C/C++**: Significant rewrite required
- **WebAssembly**: Can compile existing code, but requires Wasm toolchain

### Interoperability
- **Gopher MCP**: Best for C API integration
- **WebAssembly frameworks**: Best for multi-language composition
- **Go/Rust**: Good for new projects, limited cross-language support

---

## Conclusion

The choice of MCP framework depends on your specific requirements:

- **For simplicity and rapid development**: Go frameworks (Foxy Contexts or mcp-go)
- **For security and performance**: Rust frameworks (Turul MCP Framework)
- **For enterprise and cross-language**: C++ (Gopher MCP) or WebAssembly (Wasmcp)
- **For C/C++ codebase integration**: mcp-cpp
- **For secure sandboxing**: WebAssembly frameworks (Wassette)

Given your current Python-based MCP server, **Go frameworks** (particularly **mcp-go**) would provide the smoothest migration path while offering good performance and maintainability.

---

## Cursor IDE Integration

### Cursor MCP Protocol Support

Cursor IDE supports three transport methods for MCP servers:

1. **STDIO** - Standard input/output (local execution)
2. **SSE** - Server-Sent Events (local or remote deployment)
3. **Streamable HTTP** - HTTP-based streaming (local or remote deployment)

### Best Protocol for Cursor Integration: **STDIO** ⭐⭐⭐⭐⭐

**STDIO is the recommended protocol for Cursor integration** for the following reasons:

#### Advantages of STDIO for Cursor:

✅ **Simplest Configuration**
- Uses `command` in `mcp.json` (no URL or authentication needed)
- Cursor manages the process lifecycle automatically
- No server deployment required

✅ **Best Performance**
- Direct process communication (no network overhead)
- Lower latency than HTTP-based protocols
- Efficient for local development tools

✅ **Easiest Development**
- No need to set up HTTP servers
- No authentication/OAuth complexity
- Works out of the box with any language

✅ **Perfect for Single-User Tools**
- Ideal for personal development tools
- No multi-user concerns
- Matches Cursor's primary use case

✅ **Current Standard**
- Most MCP servers use STDIO
- Best documentation and examples
- Widely supported by all Go frameworks

#### Example STDIO Configuration for Cursor:

```json
{
  "mcpServers": {
    "my-mcp-server": {
      "command": "/path/to/server",
      "args": [],
      "env": {
        "PROJECT_ROOT": "{{PROJECT_ROOT}}"
      }
    }
  }
}
```

---

### When to Use SSE or Streamable HTTP

**Use SSE or Streamable HTTP only if:**

⚠️ **Multi-User Requirements**
- Need to share MCP server across multiple users
- Team collaboration on shared tools
- Centralized tool deployment

⚠️ **Remote Deployment**
- Server runs on different machine/network
- Cloud-based tool services
- Distributed architecture

⚠️ **OAuth Authentication**
- Need user authentication
- Secure access control
- Enterprise security requirements

**Example SSE Configuration:**

```json
{
  "mcpServers": {
    "remote-mcp-server": {
      "url": "https://mcp.example.com/sse",
      "transport": "sse"
    }
  }
}
```

---

### Go Framework Recommendations for Cursor

#### For STDIO (Recommended):

**Best Choices:**
1. **go-sdk (Official)** - Official STDIO support, guaranteed compatibility
2. **mcp-go (mark3labs)** - Excellent STDIO implementation, largest community
3. **mcp-golang (metoro-io)** - Optimized for STDIO, minimal boilerplate

**Why These Work Best:**
- All have excellent STDIO support
- Simple command-based configuration
- No HTTP server complexity
- Perfect for Cursor's `command` configuration

#### For SSE/HTTP (If Needed):

**Best Choices:**
1. **go-sdk (Official)** - Official HTTP/SSE support
2. **mcp-go (mark3labs)** - Full HTTP/SSE implementation
3. **go-mcp (ThinkInAIXYZ)** - Three-layer architecture with HTTP/SSE

---

### Migration Path: Python → Go for Cursor

**Current Setup (Python STDIO):**
```json
{
  "mcpServers": {
    "mcp-stdio-tools": {
      "command": "{{PROJECT_ROOT}}/../mcp-stdio-tools/run_server.sh"
    }
  }
}
```

**Recommended Go Migration:**

**Option 1: go-sdk (Official)**
```json
{
  "mcpServers": {
    "mcp-stdio-tools": {
      "command": "{{PROJECT_ROOT}}/../mcp-stdio-tools/bin/server"
    }
  }
}
```

**Option 2: mcp-go (mark3labs)**
```json
{
  "mcpServers": {
    "mcp-stdio-tools": {
      "command": "{{PROJECT_ROOT}}/../mcp-stdio-tools/bin/server"
    }
  }
}
```

**Benefits of Go Migration:**
- ✅ Single binary deployment (no Python dependencies)
- ✅ Faster startup time
- ✅ Lower memory footprint
- ✅ Better performance
- ✅ Same STDIO protocol (no config changes needed)

---

### Cursor Integration Checklist

**For STDIO (Recommended):**
- [x] Use `command` in `mcp.json` (not `url`)
- [x] Ensure server reads from stdin and writes to stdout
- [x] Use JSON-RPC 2.0 over stdio
- [x] Handle process lifecycle (Cursor manages it)
- [x] Support MCP protocol version 2025-06-18

**For SSE/HTTP (Advanced):**
- [ ] Set up HTTP server with SSE endpoint
- [ ] Configure OAuth if needed
- [ ] Use `url` and `transport` in `mcp.json`
- [ ] Handle authentication tokens
- [ ] Support multiple concurrent connections

---

### Summary: Best Protocol for Cursor

**🏆 Winner: STDIO**

**Why STDIO is Best:**
1. ✅ Simplest to configure and deploy
2. ✅ Best performance (no network overhead)
3. ✅ Perfect for single-user development tools
4. ✅ Cursor's primary use case
5. ✅ Works with all Go frameworks
6. ✅ No authentication complexity
7. ✅ Current industry standard

**When to Consider Alternatives:**
- Only if you need multi-user or remote deployment
- Only if OAuth authentication is required
- Only if server must run on different machine

**Recommendation:** Stick with STDIO for Cursor integration unless you have specific requirements for SSE/HTTP.

---

## References

- [Model Context Protocol Specification](https://modelcontextprotocol.io)
- [Cursor MCP Documentation](https://docs.cursor.com/context/model-context-protocol)
- [Turul MCP Framework](https://github.com/aussierobots/turul-mcp-framework)
- [mcp-cpp](https://github.com/mpsm/mcp-cpp)
- [Gopher MCP](https://github.com/GopherSecurity/gopher-mcp)
- [Wassette Blog Post](https://rye.dev/blog/wassette-webassembly-mcp-runtime/)
- [Wasmcp Blog Post](https://spinframework.dev/blog/mcp-with-wasmcp)

