// Package acp implements the Agent Client Protocol (ACP) server for exarp-go.
// When run with -acp, exarp-go speaks ACP on stdio, enabling editors like Zed,
// JetBrains, and OpenCode to use it as an AI coding agent.
package acp

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/davidl71/exarp-go/internal/framework"
	"github.com/davidl71/exarp-go/internal/logging"
	"github.com/joshgarnett/agent-client-protocol-go/acp"
	"github.com/joshgarnett/agent-client-protocol-go/acp/api"
)

const connectionTimeout = 30 * time.Second

// Server runs an ACP agent that uses the given MCP server for tools and LLM.
type Server struct {
	MCP framework.MCPServer

	agentConn     *acp.AgentConnection
	activePrompts sync.Map // sessionID -> context.CancelFunc
}

// NewServer creates an ACP server backed by the given MCP server.
func NewServer(mcp framework.MCPServer) *Server {
	return &Server{MCP: mcp}
}

// Run starts the ACP server on stdio. It blocks until the connection closes.
func (s *Server) Run(ctx context.Context) error {
	registry := acp.NewHandlerRegistry()
	registry.RegisterInitializeHandler(s.handleInitialize)
	registry.RegisterSessionNewHandler(s.handleSessionNew)
	registry.RegisterSessionPromptHandler(s.handleSessionPrompt)
	registry.RegisterSessionCancelHandler(s.handleSessionCancel)

	stdio := &stdioReadWriteCloser{Reader: os.Stdin, Writer: os.Stdout}
	conn, err := acp.NewAgentConnectionStdio(ctx, stdio, registry, connectionTimeout)
	if err != nil {
		return fmt.Errorf("create ACP connection: %w", err)
	}
	s.agentConn = conn

	logging.Warn("ACP agent started (PID: %d), waiting for client...", os.Getpid())
	return conn.Wait()
}

func (s *Server) handleInitialize(_ context.Context, params *api.InitializeRequest) (*api.InitializeResponse, error) {
	logging.Warn("[ACP] Initialize (protocol v%d)", params.ProtocolVersion)
	return &api.InitializeResponse{
		ProtocolVersion: api.ACPProtocolVersion,
		AgentCapabilities: api.AgentCapabilities{
			LoadSession:        false,
			PromptCapabilities: api.PromptCapabilities{},
		},
		AuthMethods: []api.AuthMethod{},
	}, nil
}

func (s *Server) handleSessionNew(_ context.Context, params *api.NewSessionRequest) (*api.NewSessionResponse, error) {
	sessionID := fmt.Sprintf("exarp_%d", time.Now().UnixNano())
	logging.Warn("[ACP] New session: %s (cwd=%s)", sessionID, params.Cwd)
	return &api.NewSessionResponse{
		SessionId: api.SessionId(sessionID),
	}, nil
}

func (s *Server) handleSessionPrompt(ctx context.Context, params *api.PromptRequest) (*api.PromptResponse, error) {
	logging.Warn("[ACP] Prompt for session %s", params.SessionId)

	if s.agentConn == nil {
		return &api.PromptResponse{StopReason: api.StopReasonRefusal}, fmt.Errorf("agent connection not available")
	}

	promptCtx, cancel := context.WithCancel(ctx)
	s.activePrompts.Store(string(params.SessionId), cancel)
	defer s.activePrompts.Delete(string(params.SessionId))

	stopReason, err := s.processPrompt(promptCtx, params)
	if err != nil {
		if promptCtx.Err() == context.Canceled {
			return &api.PromptResponse{StopReason: api.StopReasonCancelled}, nil
		}
		logging.Warn("[ACP] Prompt error: %v", err)
		return &api.PromptResponse{StopReason: api.StopReasonRefusal}, err
	}

	return &api.PromptResponse{StopReason: stopReason}, nil
}

func (s *Server) handleSessionCancel(_ context.Context, params *api.CancelNotification) error {
	logging.Warn("[ACP] Cancel requested for session %s", params.SessionId)
	if cancel, ok := s.activePrompts.LoadAndDelete(string(params.SessionId)); ok {
		cancel.(context.CancelFunc)()
	}
	return nil
}

func (s *Server) processPrompt(ctx context.Context, params *api.PromptRequest) (api.StopReason, error) {
	promptText := extractPromptText(params.Prompt)
	if promptText == "" {
		promptText = "(No text in prompt)"
	}

	// Call text_generate via MCP
	args := map[string]interface{}{
		"prompt":      promptText,
		"provider":    "auto",
		"max_tokens":  2048,
		"temperature": 0.7,
	}
	argsBytes, err := json.Marshal(args)
	if err != nil {
		return api.StopReasonRefusal, fmt.Errorf("marshal args: %w", err)
	}

	contents, err := s.MCP.CallTool(ctx, "text_generate", argsBytes)
	if err != nil {
		return api.StopReasonRefusal, fmt.Errorf("text_generate: %w", err)
	}

	responseText := ""
	for _, c := range contents {
		if c.Type == "text" {
			responseText += c.Text
		}
	}
	if responseText == "" {
		responseText = "(No response generated)"
	}

	// Stream response to client
	chunkSize := 200
	for i := 0; i < len(responseText); i += chunkSize {
		if ctx.Err() != nil {
			return api.StopReasonCancelled, nil
		}
		end := i + chunkSize
		if end > len(responseText) {
			end = len(responseText)
		}
		chunk := responseText[i:end]
		textContent := api.NewContentBlockText(nil, chunk)
		update := api.NewSessionUpdateAgentMessageChunk(textContent)
		if err := s.agentConn.SendSessionUpdate(ctx, &api.SessionNotification{
			SessionId: params.SessionId,
			Update:    update,
		}); err != nil {
			return api.StopReasonRefusal, fmt.Errorf("send session update: %w", err)
		}
	}

	return api.StopReasonEndTurn, nil
}

func extractPromptText(prompt []api.PromptRequestPromptElem) string {
	var parts []string
	for _, elem := range prompt {
		// JSON unmarshaling produces map[string]interface{} for objects
		if v, ok := elem.(map[string]interface{}); ok {
			if t, _ := v["type"].(string); t == "text" {
				if textObj, ok := v["text"].(map[string]interface{}); ok {
					if text, _ := textObj["text"].(string); text != "" {
						parts = append(parts, text)
					}
				}
			}
		}
	}
	return strings.TrimSpace(strings.Join(parts, "\n\n"))
}

type stdioReadWriteCloser struct {
	io.Reader
	io.Writer
}

func (s *stdioReadWriteCloser) Close() error {
	return nil
}
