.PHONY: help clean gen build run dev test fmt lint all generate-grpc

# Variables
BINARY_NAME=auth
BINARY_PATH=bin/$(BINARY_NAME)
MAIN_FILE=main.go

# Colors for output
RED=\033[0;31m
GREEN=\033[0;32m
YELLOW=\033[1;33m
BLUE=\033[0;34m
NC=\033[0m

help: ## Show this help message
	@echo "$(BLUE)Available targets:$(NC)"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(GREEN)%-15s$(NC) %s\n", $$1, $$2}'

clean: ## Clean generated files and build artifacts
	@echo "$(YELLOW)🧹 Cleaning generated files...$(NC)"
	@echo "  Removing Policy() and Hooks() from schemas..."
	@for schema_file in ent/schema/*.go; do \
		if [ -f "$$schema_file" ] && [[ ! "$$schema_file" =~ (mixin|enum) ]]; then \
			sed -i '' '/privacy ".*\/ent\/schema_privacy"/d' "$$schema_file" 2>/dev/null || true; \
			sed -i '' '/hook ".*\/ent\/schema_hooks"/d' "$$schema_file" 2>/dev/null || true; \
			sed -i '' '/^func.*Policy() ent.Policy {$$/,/^}$$/d' "$$schema_file" 2>/dev/null || true; \
			sed -i '' '/^func.*Hooks() \[\]ent.Hook {$$/,/^}$$/d' "$$schema_file" 2>/dev/null || true; \
			sed -i '' -e '/^$$/N;/^\n$$/d' "$$schema_file" 2>/dev/null || true; \
		fi; \
	done
	@rm -rf internal/ent
	@rm -rf bin/
	@echo "$(GREEN)✅ Clean complete$(NC)"

gen: clean ## Generate all code (privacy, hooks, ent, resolvers, graphql)
	@echo "$(BLUE)🔧 Generating all code...$(NC)"
	@go generate generate.go
	@echo "$(GREEN)✅ Code generation complete$(NC)"

generate-grpc: ## Generate gRPC service clients for this microservice
	@echo "$(BLUE)🔧 Generating gRPC service clients...$(NC)"
	@go run cmd/generate-grpc/main.go
	@echo "$(GREEN)✅ gRPC service clients generated$(NC)"
	@echo "$(YELLOW)💡 Run 'cd .. && make generate-clients' from project root to update service_clients.go$(NC)"

build: gen ## Build the application binary
	@echo "$(BLUE)🏗️  Building application...$(NC)"
	@mkdir -p bin
	@go build -o $(BINARY_PATH) $(MAIN_FILE)
	@echo "$(GREEN)✅ Build complete: $(BINARY_PATH)$(NC)"

run: ## Run the application (regenerates code first)
	@echo "$(BLUE)🚀 Running application...$(NC)"
	@$(MAKE) gen
	@go run $(MAIN_FILE)

dev: ## Development mode with auto-reload (requires air)
	@echo "$(BLUE)🔥 Starting development server with hot-reload...$(NC)"
	@if command -v air > /dev/null; then \
		air; \
	else \
		echo "$(RED)❌ Air not installed. Install with: go install github.com/air-verse/air@latest$(NC)"; \
		echo "$(YELLOW)💡 Falling back to regular run...$(NC)"; \
		$(MAKE) run; \
	fi

test: ## Run tests
	@echo "$(BLUE)🧪 Running tests...$(NC)"
	@go test -v ./...

fmt: ## Format code
	@echo "$(BLUE)✨ Formatting code...$(NC)"
	@gofmt -s -w .
	@go mod tidy
	@echo "$(GREEN)✅ Code formatted$(NC)"

lint: ## Run linters (requires golangci-lint)
	@echo "$(BLUE)🔍 Running linters...$(NC)"
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run ./...; \
	else \
		echo "$(YELLOW)⚠️  golangci-lint not installed$(NC)"; \
		echo "$(BLUE)💡 Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest$(NC)"; \
	fi

all: gen build test ## Run all steps: generate, build, test
	@echo "$(GREEN)✅ All tasks completed successfully!$(NC)"

quick: ## Quick run without regeneration
	@echo "$(BLUE)⚡ Quick run...$(NC)"
	@go run $(MAIN_FILE)

