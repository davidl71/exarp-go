// Package api provides an optional HTTP API for the exarp-go PWA and other clients.
// When exarp-go is started with -serve :8080, this server exposes REST endpoints
// that call the same MCP tools (task_workflow, session, report) used by the CLI.
package api

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/davidl71/exarp-go/internal/framework"
)

// Server holds the MCP server and project root for HTTP handlers.
type Server struct {
	MCPServer   framework.MCPServer
	ProjectRoot string
}

// NewServer returns an API server that proxies tool calls to the given MCPServer.
func NewServer(mcp framework.MCPServer, projectRoot string) *Server {
	return &Server{MCPServer: mcp, ProjectRoot: projectRoot}
}

type toolResponse struct {
	Contents []contentBlock `json:"contents"`
	Error    string         `json:"error,omitempty"`
}

type contentBlock struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// Handler returns the http.Handler for the API (with CORS and routing).
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/tasks", s.cors(s.handleGetTasks))
	mux.HandleFunc("GET /api/session/prime", s.cors(s.handleSessionPrime))
	mux.HandleFunc("GET /api/report/overview", s.cors(s.handleReportOverview))
	mux.HandleFunc("GET /api/report/scorecard", s.cors(s.handleReportScorecard))
	mux.HandleFunc("POST /api/tools/{name}", s.cors(s.handlePostTool))
	mux.HandleFunc("OPTIONS /api/", s.corsOptions)
	return mux
}

func (s *Server) cors(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next(w, r)
	}
}

func (s *Server) corsOptions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func (s *Server) writeToolResult(w http.ResponseWriter, contents []framework.TextContent, err error) {
	if err != nil {
		s.writeJSON(w, http.StatusInternalServerError, toolResponse{Error: err.Error()})
		return
	}
	blocks := make([]contentBlock, len(contents))
	for i, c := range contents {
		blocks[i] = contentBlock{Type: c.Type, Text: c.Text}
	}
	s.writeJSON(w, http.StatusOK, toolResponse{Contents: blocks})
}

func (s *Server) handleGetTasks(w http.ResponseWriter, r *http.Request) {
	args := map[string]interface{}{
		"action":        "sync",
		"sub_action":    "list",
		"output_format": "json",
	}
	if v := r.URL.Query().Get("status"); v != "" {
		args["status_filter"] = v
	}
	if v := r.URL.Query().Get("priority"); v != "" {
		args["priority_filter"] = v
	}
	if v := r.URL.Query().Get("tag"); v != "" {
		args["filter_tag"] = v
	}
	argsBytes, _ := json.Marshal(args)
	ctx := withProjectRoot(r.Context(), s.ProjectRoot)
	contents, err := s.MCPServer.CallTool(ctx, "task_workflow", argsBytes)
	s.writeToolResult(w, contents, err)
}

func (s *Server) handleSessionPrime(w http.ResponseWriter, r *http.Request) {
	args := map[string]interface{}{
		"action":        "prime",
		"include_tasks": true,
		"include_hints": true,
	}
	argsBytes, _ := json.Marshal(args)
	ctx := withProjectRoot(r.Context(), s.ProjectRoot)
	contents, err := s.MCPServer.CallTool(ctx, "session", argsBytes)
	s.writeToolResult(w, contents, err)
}

func (s *Server) handleReportOverview(w http.ResponseWriter, r *http.Request) {
	args := map[string]interface{}{
		"action":                  "overview",
		"include_metrics":         true,
		"include_recommendations": true,
	}
	argsBytes, _ := json.Marshal(args)
	ctx := withProjectRoot(r.Context(), s.ProjectRoot)
	contents, err := s.MCPServer.CallTool(ctx, "report", argsBytes)
	s.writeToolResult(w, contents, err)
}

func (s *Server) handleReportScorecard(w http.ResponseWriter, r *http.Request) {
	args := map[string]interface{}{
		"action":                  "scorecard",
		"include_metrics":         true,
		"include_recommendations": true,
	}
	argsBytes, _ := json.Marshal(args)
	ctx := withProjectRoot(r.Context(), s.ProjectRoot)
	contents, err := s.MCPServer.CallTool(ctx, "report", argsBytes)
	s.writeToolResult(w, contents, err)
}

func (s *Server) handlePostTool(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	if name == "" {
		s.writeJSON(w, http.StatusBadRequest, toolResponse{Error: "missing tool name"})
		return
	}
	allowed := make(map[string]bool)
	for _, t := range s.MCPServer.ListTools() {
		allowed[t.Name] = true
	}
	if !allowed[name] {
		s.writeJSON(w, http.StatusBadRequest, toolResponse{Error: "unknown tool: " + name})
		return
	}
	var args map[string]interface{}
	if r.Body != nil {
		_ = json.NewDecoder(r.Body).Decode(&args)
	}
	if args == nil {
		args = make(map[string]interface{})
	}
	argsBytes, _ := json.Marshal(args)
	ctx := withProjectRoot(r.Context(), s.ProjectRoot)
	contents, err := s.MCPServer.CallTool(ctx, name, argsBytes)
	s.writeToolResult(w, contents, err)
}

type projectRootKey struct{}

func withProjectRoot(ctx context.Context, root string) context.Context {
	return context.WithValue(ctx, projectRootKey{}, root)
}

// GetProjectRoot returns the project root from context if set by the API server.
func GetProjectRoot(ctx context.Context) (string, bool) {
	v := ctx.Value(projectRootKey{})
	if s, ok := v.(string); ok && s != "" {
		return s, true
	}
	return "", false
}
