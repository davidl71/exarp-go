// params_helpers.go — Shared helpers for MCP tool param parsing and output paths.
package tools

import (
	"path/filepath"
	"strings"

	"github.com/spf13/cast"
)

// DefaultReportOutputPath returns params["output_path"] if non-empty, else projectRoot/docs/defaultFilename.
// Use for tools that write a report and accept an optional output_path with a docs/ default.
func DefaultReportOutputPath(projectRoot, defaultFilename string, params map[string]interface{}) string {
	if p := strings.TrimSpace(cast.ToString(params["output_path"])); p != "" {
		return p
	}
	return filepath.Join(projectRoot, "docs", defaultFilename)
}
