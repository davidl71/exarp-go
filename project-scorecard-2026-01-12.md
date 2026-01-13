  📊 GO PROJECT SCORECARD
======================================================================

  OVERALL SCORE: 35.0%
  Production Ready: NO ❌

  Codebase Metrics:
    Go Files:        94
    Go Lines:        28210
    Go Test Files:   27
    Go Test Lines:   5348
    Python Files:    20 (bridge scripts)
    Python Lines:    2869
    Go Modules:      1
    Go Dependencies: 0
    Go Version:       
    MCP Tools:        24
    MCP Prompts:      15
    MCP Resources:    17

  Go Health Checks:
    go.mod exists:        ✅
    go.sum exists:        ✅
    go mod tidy:          ❌
    Go version valid:     ❌ (unknown)
    go build:             ❌
    go vet:               ❌
    go fmt:               ❌
    golangci-lint config: ✅
    golangci-lint:        ❌
    go test:              ✅
    Test coverage:        0.0%
    govulncheck:          ❌

  Security Features:
    Path boundary enforcement: ✅
    Rate limiting:             ✅
    Access control:            ✅

  Recommendations:
    • Run 'go mod tidy' to clean up dependencies
    • Fix Go build errors
    • Fix 'go vet' issues
    • Run 'go fmt ./...' to format code
    • Fix golangci-lint issues
    • Increase test coverage (currently 0.0%, target: 80%)
    • Install and run 'govulncheck ./...' for security scanning


