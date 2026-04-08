# ==============================================================================
# Makefile for hyperfleet-hooks
# ==============================================================================

.PHONY: help build build-all test test-coverage lint clean install validate-commits

# ==============================================================================
# Configuration
# ==============================================================================

# Binary configuration
BINARY_NAME := hyperfleet-hooks
BIN_DIR := bin

# Version information (auto-detected from git)
APP_VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
GIT_SHA     ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
GIT_DIRTY   ?= $(shell [ -z "$$(git status --porcelain 2>/dev/null)" ] || echo "-modified")
BUILD_DATE  ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

# Build flags
GOFLAGS ?= -trimpath
LDFLAGS := -ldflags "\
	-s -w \
	-X github.com/openshift-hyperfleet/hyperfleet-hooks/pkg/version.Version=$(APP_VERSION) \
	-X github.com/openshift-hyperfleet/hyperfleet-hooks/pkg/version.GitCommit=$(GIT_SHA)$(GIT_DIRTY) \
	-X github.com/openshift-hyperfleet/hyperfleet-hooks/pkg/version.BuildDate=$(BUILD_DATE)"

# ==============================================================================
# Targets
# ==============================================================================

.DEFAULT_GOAL := help

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# ------------------------------------------------------------------------------
# Build
# ------------------------------------------------------------------------------

build: ## Build the binary
	@echo "Building $(BINARY_NAME) $(APP_VERSION)..."
	@mkdir -p $(BIN_DIR)
	go build $(GOFLAGS) $(LDFLAGS) -o $(BIN_DIR)/$(BINARY_NAME) ./cmd/hyperfleet-hooks
	@echo "✓ Built $(BIN_DIR)/$(BINARY_NAME)"

build-all: ## Build binaries for all platforms
	@echo "Building for all platforms..."
	@mkdir -p $(BIN_DIR)
	GOOS=linux GOARCH=amd64 go build $(GOFLAGS) $(LDFLAGS) -o $(BIN_DIR)/$(BINARY_NAME)-linux-amd64 ./cmd/hyperfleet-hooks
	GOOS=linux GOARCH=arm64 go build $(GOFLAGS) $(LDFLAGS) -o $(BIN_DIR)/$(BINARY_NAME)-linux-arm64 ./cmd/hyperfleet-hooks
	GOOS=darwin GOARCH=amd64 go build $(GOFLAGS) $(LDFLAGS) -o $(BIN_DIR)/$(BINARY_NAME)-darwin-amd64 ./cmd/hyperfleet-hooks
	GOOS=darwin GOARCH=arm64 go build $(GOFLAGS) $(LDFLAGS) -o $(BIN_DIR)/$(BINARY_NAME)-darwin-arm64 ./cmd/hyperfleet-hooks
	@echo "✓ Built all platform binaries"

# ------------------------------------------------------------------------------
# Test
# ------------------------------------------------------------------------------

test: ## Run tests
	@echo "Running tests..."
	go test -v -race -coverprofile=coverage.out ./...
	@echo "✓ Tests passed"

test-coverage: test ## Run tests with coverage report
	@echo "Generating coverage report..."
	go tool cover -html=coverage.out -o coverage.html
	@echo "✓ Coverage report generated: coverage.html"

# ------------------------------------------------------------------------------
# Quality
# ------------------------------------------------------------------------------

lint: ## Run linters
	@echo "Running linters..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run --timeout=5m; \
	else \
		echo "⚠ golangci-lint not installed, running go fmt only"; \
		test -z "$$(gofmt -l .)" || (echo "Files need formatting:" && gofmt -l . && exit 1); \
	fi
	@echo "✓ Linting passed"

# ------------------------------------------------------------------------------
# Utility
# ------------------------------------------------------------------------------

clean: ## Clean build artifacts
	@echo "Cleaning..."
	rm -rf $(BIN_DIR)
	rm -f coverage.out coverage.html
	@echo "✓ Cleaned"

install: build ## Install binary to /usr/local/bin
	@echo "Installing $(BINARY_NAME) to /usr/local/bin..."
	@cp $(BIN_DIR)/$(BINARY_NAME) /usr/local/bin/$(BINARY_NAME)
	@echo "✓ Installed $(BINARY_NAME)"

# ------------------------------------------------------------------------------
# CI
# ------------------------------------------------------------------------------

validate-commits: build ## Validate commits in current branch (CI mode)
	@echo "Validating commits..."
	@./bin/hyperfleet-hooks commitlint --pr
