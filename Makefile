# Check if .env exists and include it
ifeq (,$(wildcard ./.env))
    $(error .env file not found. Please create one.)
endif

include .env
export

all: build test

build:
	@echo "Building..."
	
	
	@go build -o main apps/backend/cmd/api/main.go

run:
	@go run apps/backend/cmd/api/main.go

# Create DB container
docker-run:
	@if docker compose up --build 2>/dev/null; then \
		: ; \
	else \
		echo "Falling back to Docker Compose V1"; \
		docker-compose up --build; \
	fi

# Shutdown DB container
docker-down:
	@if docker compose down 2>/dev/null; then \
		: ; \
	else \
		echo "Falling back to Docker Compose V1"; \
		docker-compose down; \
	fi

# Test the application
test:
	@echo "Testing..."
	@go test ./... -v

# Transport Tests - All transport layer tests
test-transport:
	@echo "Running transport tests..."
	@./apps/backend/tests/run_tests.sh transport

# Integration Tests - All integration tests
test-integration:
	@echo "Running integration tests..."
	@./apps/backend/tests/run_tests.sh integration

# Unit Tests - All unit tests
test-unit:
	@echo "Running unit tests..."
	@./apps/backend/tests/run_tests.sh unit

# Specific transport tests
test-rpc:
	@echo "Running JSON-RPC transport tests..."
	@./apps/backend/tests/run_tests.sh rpc

test-sse:
	@echo "Running SSE transport tests..."
	@./apps/backend/tests/run_tests.sh sse

test-websocket:
	@echo "Running WebSocket transport tests..."
	@./apps/backend/tests/run_tests.sh websocket

test-mcp:
	@echo "Running MCP transport tests..."
	@./apps/backend/tests/run_tests.sh mcp

test-stdio:
	@echo "Running STDIO transport tests..."
	@./apps/backend/tests/run_tests.sh stdio

# Test with coverage
test-coverage:
	@echo "Running tests with coverage..."
	@cd apps/backend && TEST_COVERAGE=true ./tests/run_tests.sh all

# Test with verbose output
test-verbose:
	@echo "Running tests with verbose output..."
	@cd apps/backend && TEST_VERBOSE=true ./tests/run_tests.sh all

# Legacy integration test (for database)
itest:
	@echo "Running database integration tests..."
	@go test ./apps/backend/internal/database -v

# Run all transport tests
test-all-transports: test-rpc test-sse test-websocket test-mcp test-stdio
	@echo "All transport tests completed!"

# Clean the binary
clean:
	@echo "Cleaning..."
	@rm -f main api apps/backend/api
	@rm -f logs/*.log

# Live Reload
watch:
	@if command -v air > /dev/null; then \
            air; \
            echo "Watching...";\
        else \
            read -p "Go's 'air' is not installed on your machine. Do you want to install it? [Y/n] " choice; \
            if [ "$$choice" != "n" ] && [ "$$choice" != "N" ]; then \
                go install github.com/air-verse/air@latest; \
                air; \
                echo "Watching...";\
            else \
                echo "You chose not to install air. Exiting..."; \
                exit 1; \
            fi; \
        fi

# Run database migrations
migrate:
	@echo "Running database migrations..."
	@bash -c 'cd apps/backend && migrate -path migrations -database "postgres://$$DB_USERNAME:$$DB_PASSWORD@$$DB_HOST:$$DB_PORT/$$DB_DATABASE?sslmode=disable&search_path=$$DB_SCHEMA" up'

# Rollback database migrations
migrate-down:
	@echo "Rolling back database migrations..."
	@bash -c 'cd apps/backend && migrate -path migrations -database "postgres://$$DB_USERNAME:$$DB_PASSWORD@$$DB_HOST:$$DB_PORT/$$DB_DATABASE?sslmode=disable&search_path=$$DB_SCHEMA" down'

# Check migration status
migrate-status:
	@echo "Checking migration status..."
	@bash -c 'cd apps/backend && migrate -path migrations -database "postgres://$$DB_USERNAME:$$DB_PASSWORD@$$DB_HOST:$$DB_PORT/$$DB_DATABASE?sslmode=disable&search_path=$$DB_SCHEMA" status'

# Create new migration file
migrate-create:
	@echo "Creating new migration file..."
	@read -p "Enter migration name: " name; \
	cd apps/backend && migrate create -ext sql -dir migrations -seq $$name

# Setup commands for local development
setup:
	@echo "Running local development setup..."
	@go run apps/backend/cmd/setup/main.go

setup-reset:
	@echo "Resetting database..."
	@go run apps/backend/cmd/setup/main.go reset

# Pre-commit hooks setup
setup-precommit:
	@echo "Installing pre-commit hooks..."
	@if command -v pre-commit > /dev/null; then \
		pre-commit install; \
		echo "Pre-commit hooks installed successfully!"; \
	else \
		echo "Installing pre-commit..."; \
		pip install pre-commit; \
		pre-commit install; \
		echo "Pre-commit hooks installed successfully!"; \
	fi

# Run pre-commit on all files
precommit-all:
	@echo "Running pre-commit on all files..."
	@pre-commit run --all-files

# Code quality checks
lint:
	@echo "Running golangci-lint..."
	@golangci-lint run

lint-fix:
	@echo "Running golangci-lint with fixes..."
	@golangci-lint run --fix

security:
	@echo "Running security checks..."
	@gosec ./...

.PHONY: all build run test clean watch docker-run docker-down itest migrate migrate-down migrate-status migrate-create test-transport test-integration test-unit test-rpc test-sse test-websocket test-mcp test-stdio test-coverage test-verbose test-all-transports setup setup-admin setup-org setup-dummy setup-reset setup-precommit precommit-all lint lint-fix security