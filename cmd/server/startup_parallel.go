// startup_parallel.go — Overlap project-root/database setup with MCP server construction.
//
// FindProjectRoot and EnsureConfigAndDatabase are I/O heavy; config.Load and NewServerFromConfig
// are independent. Registration of tools, prompts, and resources must stay sequential: the Go SDK
// adapter mutates internal maps without locking (see mcp-go-core gosdk adapter RegisterTool).
package main

import (
	"sync"

	"github.com/davidl71/exarp-go/internal/cli"
	"github.com/davidl71/exarp-go/internal/config"
	"github.com/davidl71/exarp-go/internal/factory"
	"github.com/davidl71/exarp-go/internal/framework"
	toolsx "github.com/davidl71/exarp-go/internal/tools"
)

// loadDatabaseAndServerInParallel runs EnsureConfigAndDatabase concurrently with config.Load +
// NewServerFromConfig. Returns fatalErr if the server cannot be constructed; rootErr if the
// project root could not be resolved (non-fatal for some modes).
func loadDatabaseAndServerInParallel(toolFilter framework.ToolFilterFunc) (server framework.MCPServer, projectRoot string, rootErr error, fatalErr error) {
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		projectRoot, rootErr = toolsx.FindProjectRoot()
		if rootErr == nil {
			cli.EnsureConfigAndDatabase(projectRoot)
		}
	}()
	go func() {
		defer wg.Done()
		cfg, err := config.Load()
		if err != nil {
			fatalErr = err
			return
		}
		server, fatalErr = factory.NewServerFromConfig(cfg, factory.WithToolFilter(toolFilter))
	}()
	wg.Wait()
	return server, projectRoot, rootErr, fatalErr
}
