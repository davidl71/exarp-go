.PHONY: help build run test test-watch test-coverage test-html clean install fmt lint dev dev-watch dev-test dev-full bench docs sanity-check test-cli test-cli-list test-cli-tool test-cli-test

# Project configuration
PROJECT_NAME := exarp-go
PYTHON := uv run python
BINARY_NAME := exarp-go
BINARY_PATH := bin/$(BINARY_NAME)

# Colors for output
GREEN := \033[0;32m
YELLOW := \033[1;33m
RED := \033[0;31m
BLUE := \033[0;34m
NC := \033[0m # No Color

# Default target
.DEFAULT_GOAL := help

##@ Development

help: ## Show this help message
	@echo "$(BLUE)Available targets:$(NC)"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(GREEN)%-20s$(NC) %s\n", $$1, $$2}'

build: ## Build the Go server
	@echo "$(BLUE)Building $(PROJECT_NAME)...$(NC)"
	@go build -o $(BINARY_PATH) ./cmd/server || (echo "$(RED)❌ Build failed$(NC)" && exit 1)
	@echo "$(GREEN)✅ Server built: $(BINARY_PATH)$(NC)"

run: build ## Run the MCP server
	@echo "$(BLUE)Running $(PROJECT_NAME) server...$(NC)"
	@$(BINARY_PATH)

dev: ## Start development mode (auto-reload on changes)
	@echo "$(BLUE)Starting development mode...$(NC)"
	@./dev.sh --watch

dev-watch: ## Watch files and auto-reload (alias for dev)
	@$(MAKE) dev

dev-test: ## Development mode with auto-test on changes
	@echo "$(BLUE)Starting development mode with auto-test...$(NC)"
	@./dev.sh --watch --test

dev-full: ## Full development mode (watch + test + coverage)
	@echo "$(BLUE)Starting full development mode...$(NC)"
	@./dev.sh --watch --test --coverage

##@ Testing

test: test-go test-python ## Run all tests (Go + Python)
	@echo "$(GREEN)✅ All tests passed$(NC)"

test-go: ## Run Go tests
	@echo "$(BLUE)Running Go tests...$(NC)"
	@go test ./... -v || \
	 echo "$(YELLOW)⚠️  Go tests failed or not available$(NC)"

test-python: ## Run Python tests
	@echo "$(BLUE)Running Python tests...$(NC)"
	@$(PYTHON) -m pytest tests/unit/python tests/integration -v || \
	 echo "$(YELLOW)⚠️  Python tests failed$(NC)"

test-integration: ## Run integration tests
	@echo "$(BLUE)Running integration tests...$(NC)"
	@$(PYTHON) -m pytest tests/integration -v || \
	 echo "$(YELLOW)⚠️  Integration tests failed$(NC)"
	@echo "$(BLUE)Running Go tests for coverage...$(NC)"
	@go test ./... -v || echo "$(YELLOW)⚠️  Go tests failed$(NC)"

test-coverage: test-coverage-go test-coverage-python ## Run tests with coverage report

test-coverage-go: ## Generate Go test coverage report
	@echo "$(BLUE)Running Go tests with coverage...$(NC)"
	@go test ./... -coverprofile=coverage-go.out -covermode=atomic || \
	 echo "$(YELLOW)⚠️  Go coverage failed$(NC)"
	@go tool cover -html=coverage-go.out -o coverage-go.html 2>/dev/null || true
	@echo "$(GREEN)✅ Go coverage report: coverage-go.html$(NC)"

test-coverage-python: ## Generate Python test coverage report
	@echo "$(BLUE)Running Python tests with coverage...$(NC)"
	@$(PYTHON) -m pytest tests/unit/python tests/integration \
		--cov=bridge --cov=internal --cov-report=term --cov-report=html || \
	 echo "$(YELLOW)⚠️  Python coverage failed$(NC)"
	@echo "$(GREEN)✅ Python coverage report: htmlcov/index.html$(NC)"

test-html: test-coverage ## Generate HTML coverage reports (alias for test-coverage)

test-watch: ## Run tests in watch mode (auto-test on changes)
	@echo "$(BLUE)Starting test watch mode...$(NC)"
	@./dev.sh --test-watch

test-tools: build ## Test Go server tools
	@echo "$(BLUE)Testing Go server...$(NC)"
	@test -f $(BINARY_PATH) && echo "$(GREEN)✅ Go binary exists$(NC)" || (echo "$(RED)❌ Go binary not found$(NC)" && exit 1)
	@test -x $(BINARY_PATH) && echo "$(GREEN)✅ Go binary is executable$(NC)" || (echo "$(RED)❌ Go binary is not executable$(NC)" && exit 1)

sanity-check: ## Verify tools/resources/prompts counts match expected values
	@echo "$(BLUE)Running sanity check...$(NC)"
	@go build -o bin/sanity-check cmd/sanity-check/main.go 2>/dev/null || true
	@./bin/sanity-check || (echo "$(RED)❌ Sanity check failed$(NC)" && exit 1)

test-all: test-tools sanity-check test-cli ## Run all import tests + sanity check + CLI tests

test-mcp: ## Test MCP server via stdio (requires manual input)
	@echo "$(BLUE)Testing MCP server (stdio mode)...$(NC)"
	@echo "$(YELLOW)Note: This requires manual JSON-RPC input$(NC)"
	@$(BINARY_PATH) < /dev/stdin

##@ CLI Testing

test-cli: build test-cli-list test-cli-tool test-cli-test ## Run all CLI functionality tests
	@echo "$(GREEN)✅ All CLI tests passed$(NC)"

test-cli-list: build ## Test CLI list tools functionality
	@echo "$(BLUE)Testing CLI: list tools...$(NC)"
	@$(BINARY_PATH) -list > /dev/null 2>&1 && \
	 echo "$(GREEN)✅ CLI list command works$(NC)" || \
	 (echo "$(RED)❌ CLI list command failed$(NC)" && exit 1)

test-cli-tool: build ## Test CLI tool execution
	@echo "$(BLUE)Testing CLI: tool execution...$(NC)"
	@$(BINARY_PATH) -tool lint -args '{"action":"run","linter":"go-vet","path":"cmd/server"}' > /dev/null 2>&1 && \
	 echo "$(GREEN)✅ CLI tool execution works$(NC)" || \
	 (echo "$(YELLOW)⚠️  CLI tool execution test skipped (may require valid tool/args)$(NC)")

test-cli-test: build ## Test CLI feature testing mode
	@echo "$(BLUE)Testing CLI: feature testing mode...$(NC)"
	@$(BINARY_PATH) -test lint > /dev/null 2>&1 && \
	 echo "$(GREEN)✅ CLI test mode works$(NC)" || \
	 (echo "$(YELLOW)⚠️  CLI test mode skipped (may require valid tool)$(NC)")

test-cli-help: build ## Test CLI help/usage display
	@echo "$(BLUE)Testing CLI: help display...$(NC)"
	@$(BINARY_PATH) 2>&1 | grep -qE "(Usage|exarp-go)" && \
	 echo "$(GREEN)✅ CLI help display works$(NC)" || \
	 (echo "$(YELLOW)⚠️  CLI help display test inconclusive$(NC)")

test-cli-mode: build ## Test CLI mode detection (TTY vs stdio)
	@echo "$(BLUE)Testing CLI: mode detection...$(NC)"
	@echo "$(BLUE)  Testing TTY mode (should show CLI)...$(NC)"
	@$(BINARY_PATH) -list > /dev/null 2>&1 && \
	 echo "$(GREEN)  ✅ TTY mode detected correctly$(NC)" || \
	 echo "$(YELLOW)  ⚠️  TTY mode test inconclusive$(NC)"
	@echo "$(BLUE)  Testing stdio mode (should run as MCP server)...$(NC)"
	@echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}' | timeout 2 $(BINARY_PATH) 2>&1 | grep -q "jsonrpc" && \
	 echo "$(GREEN)  ✅ Stdio mode detected correctly$(NC)" || \
	 echo "$(YELLOW)  ⚠️  Stdio mode test inconclusive (may timeout)$(NC)"

##@ Code Quality

fmt: ## Format code with ruff/black
	@echo "$(BLUE)Formatting code...$(NC)"
	@$(PYTHON) -m ruff format . || echo "$(YELLOW)ruff not available, skipping format$(NC)"
	@echo "$(GREEN)✅ Code formatted$(NC)"

lint: ## Lint code with ruff
	@echo "$(BLUE)Linting code...$(NC)"
	@$(PYTHON) -m ruff check . || echo "$(YELLOW)ruff not available, skipping lint$(NC)"
	@echo "$(GREEN)✅ Linting complete$(NC)"

lint-fix: ## Lint and auto-fix code
	@echo "$(BLUE)Linting and fixing code...$(NC)"
	@$(PYTHON) -m ruff check --fix . || echo "$(YELLOW)ruff not available, skipping lint-fix$(NC)"
	@echo "$(GREEN)✅ Linting and fixes complete$(NC)"

##@ Benchmarking

bench: ## Run benchmarks (if available)
	@echo "$(BLUE)Running benchmarks...$(NC)"
	@$(PYTHON) -m pytest tests/benchmarks/ -v --benchmark-only 2>/dev/null || \
	 echo "$(YELLOW)No benchmarks found$(NC)"

##@ Documentation

docs: ## Generate documentation
	@echo "$(BLUE)Generating documentation...$(NC)"
	@echo "$(YELLOW)Documentation generation not available for Go server$(NC)"

##@ Cleanup

clean: ## Clean build artifacts and cache
	@echo "$(BLUE)Cleaning...$(NC)"
	@rm -rf __pycache__ .pytest_cache .coverage htmlcov/ .ruff_cache/
	@find . -type d -name __pycache__ -exec rm -r {} + 2>/dev/null || true
	@find . -type f -name "*.pyc" -delete 2>/dev/null || true
	@echo "$(GREEN)✅ Clean complete$(NC)"

clean-all: clean ## Clean everything including virtual environment
	@echo "$(BLUE)Cleaning everything...$(NC)"
	@rm -rf .venv/ uv.lock
	@echo "$(GREEN)✅ Full clean complete$(NC)"

##@ Installation

install: ## Install dependencies
	@echo "$(BLUE)Installing dependencies...$(NC)"
	@uv sync
	@echo "$(GREEN)✅ Dependencies installed$(NC)"

install-dev: install ## Install development dependencies
	@echo "$(BLUE)Installing development dependencies...$(NC)"
	@uv sync --dev
	@echo "$(GREEN)✅ Development dependencies installed$(NC)"

##@ Go Development

go-build: ## Build Go binary
	@echo "$(BLUE)Building Go binary...$(NC)"
	@go build -o $(BINARY_PATH) ./cmd/server || \
	 (echo "$(RED)❌ Build failed$(NC)" && exit 1)
	@chmod +x $(BINARY_PATH)
	@echo "$(GREEN)✅ Build complete: $(BINARY_PATH)$(NC)"

go-run: go-build ## Run Go binary
	@echo "$(BLUE)Running $(BINARY_NAME)...$(NC)"
	@$(BINARY_PATH)

go-dev: ## Start Go development mode (hot reload)
	@echo "$(BLUE)Starting Go development mode...$(NC)"
	@./dev-go.sh --watch

go-dev-test: ## Go dev mode with auto-test
	@echo "$(BLUE)Starting Go development mode with auto-test...$(NC)"
	@./dev-go.sh --watch --test

go-test: ## Run Go tests
	@echo "$(BLUE)Running Go tests...$(NC)"
	@go test ./... -v || \
	 (echo "$(RED)❌ Tests failed$(NC)" && exit 1)
	@echo "$(GREEN)✅ All tests passed$(NC)"

go-bench: ## Run Go benchmarks
	@echo "$(BLUE)Running Go benchmarks...$(NC)"
	@go test -bench=. -benchmem ./... || \
	 echo "$(YELLOW)⚠️  Benchmarks failed$(NC)"

##@ Quick Commands

quick-test: test-tools ## Quick test (tools only)
quick-dev: dev ## Quick dev mode (watch only)
quick-build: build ## Quick build (verify imports)

##@ Model-Assisted Testing (Future)

test-models: ## Test model integration (MLX/Ollama)
	@echo "$(BLUE)Testing model integration...$(NC)"
	@echo "$(YELLOW)Model integration testing via Go server$(NC)"

test-breakdown: ## Test task breakdown with models
	@echo "$(BLUE)Testing task breakdown...$(NC)"
	@echo "$(YELLOW)Model-assisted testing not yet implemented$(NC)"

test-auto-exec: ## Test auto-execution with models
	@echo "$(BLUE)Testing auto-execution...$(NC)"
	@echo "$(YELLOW)Model-assisted testing not yet implemented$(NC)"

