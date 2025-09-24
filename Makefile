# meow meow API - Makefile

# Go parameters
GOCMD=GOTOOLCHAIN=go1.24.6 go
GOBUILD=$(GOCMD) build
GOMOD=$(GOCMD) mod
BINARY_NAME=meow
MAIN_PATH=./cmd/server
BIN_DIR=bin

.PHONY: help build run tidy up down docs clean kill logs check fmt vet test lint all

# Default target
.DEFAULT_GOAL := help

# Show help
help: ## Show this help message
	@echo "🐱 meow meow API - Available Commands"
	@echo "=========================================="
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-12s\033[0m %s\n", $$1, $$2}'
	@echo ""
	@echo "📋 Examples:"
	@echo "  make all              # Complete build pipeline (recommended)"
	@echo "  make check            # Quick quality checks (fmt, vet, test, build)"
	@echo "  make up && make run   # Start services and run app"
	@echo "  make kill             # Kill app if stuck running"
	@echo "  make logs             # Analyze application logs"
	@echo "  make down             # Stop all services"

build: ## Build the application binary
	mkdir -p $(BIN_DIR)
	$(GOBUILD) -o $(BIN_DIR)/$(BINARY_NAME) $(MAIN_PATH)

run: ## Run the application in development mode
	$(GOCMD) run $(MAIN_PATH)

tidy: ## Download and organize Go dependencies
	$(GOMOD) tidy

up: ## Start all services (PostgreSQL, Redis, MinIO) with Docker Compose
	docker compose up -d

down: ## Stop all Docker Compose services and remove volumes
	docker compose down -v

chatwoot-up: ## Start Chatwoot development environment
	docker compose -f chatwoot-dev.yml up -d

chatwoot-down: ## Stop Chatwoot development environment
	docker compose -f chatwoot-dev.yml down -v

chatwoot-logs: ## Show Chatwoot logs
	docker compose -f chatwoot-dev.yml logs -f

chatwoot-restart: ## Restart Chatwoot services
	docker compose -f chatwoot-dev.yml restart

chatwoot-status: ## Show Chatwoot services status
	docker compose -f chatwoot-dev.yml ps

full-up: up chatwoot-up ## Start both zpmeow and Chatwoot environments

full-down: down chatwoot-down ## Stop both zpmeow and Chatwoot environments

docs: ## Generate Swagger API documentation
	@if command -v swag > /dev/null; then \
		swag init -g $(MAIN_PATH)/main.go -o ./docs 2>&1 | grep -v "warning: failed to get package name" | grep -v "warning: failed to evaluate const"; \
	else \
		echo "Swagger not found. Install with: go install github.com/swaggo/swag/cmd/swag@latest"; \
	fi

clean: ## Clean build files and binaries
	rm -rf $(BIN_DIR)

kill: ## Kill any process running on port 8080
	@echo "🔍 Looking for processes on port 8080..."
	@PID=$$(lsof -ti:8080); \
	if [ -n "$$PID" ]; then \
		echo "💀 Killing process(es): $$PID"; \
		kill -9 $$PID; \
		echo "✅ Process(es) killed successfully"; \
	else \
		echo "ℹ️  No process found running on port 8080"; \
	fi

logs: ## Analyze application logs
	@echo "📊 meow Log Analysis"
	@echo "====================="
	@if [ -f "log/app.log" ]; then \
		echo "📁 Log file: log/app.log"; \
		echo "📏 Total lines: $$(wc -l < log/app.log)"; \
		echo "📅 Date range: $$(head -1 log/app.log | jq -r .time) to $$(tail -1 log/app.log | jq -r .time)"; \
		echo ""; \
		echo "🔍 Log levels:"; \
		grep -o '"level":"[^"]*"' log/app.log | sort | uniq -c | sort -nr; \
		echo ""; \
		echo "🏷️  Modules:"; \
		grep -o '"module":"[^"]*"' log/app.log | sort | uniq -c | sort -nr | head -10; \
		echo ""; \
		echo "❌ Recent errors:"; \
		grep '"level":"error"' log/app.log | tail -5 | jq -r '.time + " " + .message'; \
		echo ""; \
		echo "⚠️  Recent warnings:"; \
		grep '"level":"warn"' log/app.log | tail -5 | jq -r '.time + " " + .message'; \
	else \
		echo "❌ Log file not found: log/app.log"; \
	fi

fmt: ## Format Go code using gofmt
	@echo "🎨 Formatting Go code..."
	@gofmt -s -w .
	@echo "✅ Code formatting completed"

vet: ## Run go vet to examine Go source code
	@echo "🔍 Running go vet..."
	@$(GOCMD) vet ./...
	@echo "✅ go vet completed"

test: ## Run all tests
	@echo "🧪 Running tests..."
	@$(GOCMD) test -v ./...
	@echo "✅ Tests completed"

lint: ## Run golangci-lint (install if not present)
	@echo "🔎 Running linter..."
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run --timeout=5m; \
	else \
		echo "⚠️  golangci-lint not found. Installing..."; \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
		golangci-lint run --timeout=5m; \
	fi
	@echo "✅ Linting completed"

check: fmt vet lint test build ## Run all code quality checks (format, vet, lint, test, build)
	@echo "✅ All checks completed successfully!"

all: tidy fmt vet lint test build docs ## Complete build pipeline (tidy, format, vet, lint, test, build, docs)
	@echo "🎉 Complete build pipeline finished successfully!"
	@echo "📦 Binary available at: $(BIN_DIR)/$(BINARY_NAME)"
