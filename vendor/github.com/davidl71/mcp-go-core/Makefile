.PHONY: help build test clean fmt lint install-tools

# Project configuration
PROJECT_NAME := mcp-go-core
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Colors for output
GREEN := \033[0;32m
YELLOW := \033[1;33m
RED := \033[0;31m
BLUE := \033[0;34m
NC := \033[0m # No Color

# Default target
.DEFAULT_GOAL := help

##@ Main Targets

help: ## Show this help message
	@echo "$(BLUE)Available targets:$(NC)"
	@grep -hE '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "} {printf "  $(GREEN)%-20s$(NC) %s\n", $$1, $$2}'
	@echo ""
	@echo "$(BLUE)Project:$(NC) $(PROJECT_NAME)"
	@echo "$(BLUE)Version:$(NC) $(VERSION)"

##@ Development

build: ## Build the library (verify compilation)
	@echo "$(BLUE)Building $(PROJECT_NAME)...$(NC)"
	@go build ./...
	@echo "$(GREEN)✅ Build successful$(NC)"

test: ## Run tests
	@echo "$(BLUE)Running tests...$(NC)"
	@go test ./... -v
	@echo "$(GREEN)✅ Tests passed$(NC)"

test-coverage: ## Run tests with coverage
	@echo "$(BLUE)Running tests with coverage...$(NC)"
	@go test ./... -coverprofile=coverage.out
	@go tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)✅ Coverage report: coverage.html$(NC)"

##@ Code Quality

fmt: ## Format code
	@echo "$(BLUE)Formatting code...$(NC)"
	@go fmt ./...
	@echo "$(GREEN)✅ Code formatted$(NC)"

lint: ## Lint code
	@echo "$(BLUE)Linting code...$(NC)"
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run ./... || (echo "$(RED)❌ Linting failed$(NC)" && exit 1); \
		echo "$(GREEN)✅ Linting passed$(NC)"; \
	else \
		echo "$(YELLOW)⚠️  golangci-lint not found$(NC)"; \
		echo "$(YELLOW)   Install: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest$(NC)"; \
		exit 1; \
	fi

vet: ## Run go vet
	@echo "$(BLUE)Running go vet...$(NC)"
	@go vet ./...
	@echo "$(GREEN)✅ go vet passed$(NC)"

check: fmt vet lint ## Check code quality (fmt + vet + lint)
	@echo "$(GREEN)✅ All checks passed$(NC)"

##@ Cleanup

clean: ## Clean build artifacts
	@echo "$(BLUE)Cleaning...$(NC)"
	@rm -f coverage.out coverage.html
	@go clean
	@echo "$(GREEN)✅ Clean complete$(NC)"

##@ Installation

install-tools: ## Install Go development tools
	@echo "$(BLUE)Installing Go development tools...$(NC)"
	@if ! command -v golangci-lint >/dev/null 2>&1; then \
		echo "$(BLUE)Installing golangci-lint...$(NC)"; \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest || \
		(echo "$(YELLOW)⚠️  Failed to install golangci-lint$(NC)" && exit 1); \
		echo "$(GREEN)✅ golangci-lint installed$(NC)"; \
	else \
		echo "$(GREEN)✅ golangci-lint already installed$(NC)"; \
	fi

##@ Module Management

mod-tidy: ## Run go mod tidy
	@echo "$(BLUE)Running go mod tidy...$(NC)"
	@go mod tidy
	@echo "$(GREEN)✅ Dependencies cleaned$(NC)"

mod-verify: ## Verify go.mod and go.sum
	@echo "$(BLUE)Verifying go.mod and go.sum...$(NC)"
	@go mod verify
	@echo "$(GREEN)✅ Module verification passed$(NC)"

##@ Version

version: ## Show version information
	@echo "$(BLUE)Version Information:$(NC)"
	@echo "  Project:     $(PROJECT_NAME)"
	@echo "  Version:     $(VERSION)"
	@echo "  Build Time:  $(BUILD_TIME)"
	@echo "  Git Commit:  $(GIT_COMMIT)"
	@echo "  Go Version:  $$(go version 2>/dev/null || echo 'unknown')"
