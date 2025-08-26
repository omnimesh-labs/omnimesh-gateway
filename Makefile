# MCP Gateway Makefile
.PHONY: help dev stop clean test migrate lint setup shell bash migrate-down migrate-status setup-admin logs nuclear restart prune docker-prune docker-reset

# Docker compose command
DOCKER_COMPOSE = docker compose
BACKEND = $(DOCKER_COMPOSE) run --rm backend
GO = $(BACKEND) go

# Default target - show help
help:
	@echo "MCP Gateway - Available Commands"
	@echo ""
	@echo "Quick Start:"
	@echo "  make dev          - Start full stack with hot reload"
	@echo "  make stop         - Stop all services (clean)"
	@echo "  make restart      - Clean restart services"
	@echo "  make clean        - Stop and remove all data"
	@echo ""
	@echo "Docker Management:"
	@echo "  make prune        - Clean unused Docker resources (safe)"
	@echo "  make docker-prune - Aggressive Docker cleanup"
	@echo "  make docker-reset - Reset all project Docker resources"
	@echo "  make rebuild      - Force rebuild containers"
	@echo ""
	@echo "Database:"
	@echo "  make migrate      - Run database migrations"
	@echo "  make migrate-down - Rollback migrations"
	@echo "  make migrate-status - Show migration status"
	@echo "  make db-shell     - Open PostgreSQL shell"
	@echo ""
	@echo "Testing:"
	@echo "  make test         - Run all tests"
	@echo "  make test-transport - Transport layer tests"
	@echo "  make test-integration - Integration tests"
	@echo "  make test-unit    - Unit tests"
	@echo ""
	@echo "Development:"
	@echo "  make shell        - Open shell in backend container"
	@echo "  make bash         - Open bash in backend container"
	@echo "  make logs         - View service logs"
	@echo "  make lint         - Run linters"
	@echo "  make watch        - Local development with hot reload"
	@echo ""
	@echo "Setup:"
	@echo "  make setup        - Initial project setup"
	@echo "  make setup-admin  - Create admin user"
	@echo ""
	@echo "Memory/Disk Management:"
	@echo "  If running out of space, try these in order:"
	@echo "  1. make restart   - Clean restart (safest)"
	@echo "  2. make prune     - Remove unused resources"
	@echo "  3. make clean     - Full cleanup with data loss"
	@echo "  4. make docker-prune - Aggressive system cleanup"

# Development mode with hot reload
dev:
	@echo "Starting MCP Gateway Stack with hot reload..."
	@if [ ! -f .env ]; then \
		echo "Creating .env file from .env.example..."; \
		cp .env.example .env; \
		echo ".env file created"; \
	fi
	@$(DOCKER_COMPOSE) up

# Stop all services
stop:
	@echo "Stopping MCP Gateway Stack..."
	@$(DOCKER_COMPOSE) down --remove-orphans
	@echo "All services stopped"

# Clean everything (including volumes)
clean:
	@echo "Cleaning MCP Gateway Stack (removes all data)..."
	@$(DOCKER_COMPOSE) down -v --remove-orphans --rmi local
	@docker system prune -f
	@rm -f main api apps/backend/api logs/*.log tmp/main
	@echo "All services stopped and data removed"

# View logs
logs:
	@$(DOCKER_COMPOSE) logs -f

# Run tests
test:
	@echo "Running tests..."
	@$(GO) test -v ./...

# Run specific tests
test-transport:
	@echo "Running transport tests..."
	@$(GO) test -v ./apps/backend/tests/transport/...

test-integration:
	@echo "Running integration tests..."
	@$(GO) test -v ./apps/backend/tests/integration/...

test-unit:
	@echo "Running unit tests..."
	@$(GO) test -v ./apps/backend/tests/unit/...

# Database operations
migrate:
	@echo "Running database migrations..."
	@$(DOCKER_COMPOSE) run --rm --entrypoint /app/migrate backend up

migrate-down:
	@echo "Rolling back migrations..."
	@$(DOCKER_COMPOSE) run --rm --entrypoint /app/migrate backend down

migrate-status:
	@echo "Checking migration status..."
	@$(DOCKER_COMPOSE) run --rm --entrypoint /app/migrate backend status

# Create admin user
setup-admin:
	@echo "Setting up admin user..."
	@$(DOCKER_COMPOSE) run --rm --entrypoint /app/migrate backend setup-admin

# Development helpers
shell:
	@echo "Opening shell in backend container..."
	@$(DOCKER_COMPOSE) run --rm backend sh

bash:
	@echo "Opening bash in backend container..."
	@$(DOCKER_COMPOSE) run --rm backend bash

# Run linters
lint:
	@echo "Running linters..."
	@$(GO) run github.com/golangci/golangci-lint/cmd/golangci-lint@latest run

# Initial project setup
setup:
	@echo "Setting up project..."
	@if [ ! -f .env ]; then \
		cp .env.example .env; \
		echo "Created .env file"; \
	fi
	@echo "Installing pre-commit hooks..."
	@if command -v pre-commit > /dev/null; then \
		pre-commit install; \
	else \
		echo "Warning: pre-commit not installed. Install with: pip install pre-commit"; \
	fi
	@echo "Setup complete! Run 'make dev' to start development"

# Build only (no run)
build:
	@echo "Building containers..."
	@$(DOCKER_COMPOSE) build

# Rebuild and restart
rebuild:
	@echo "Rebuilding containers..."
	@$(DOCKER_COMPOSE) down --remove-orphans
	@$(DOCKER_COMPOSE) build --no-cache
	@$(DOCKER_COMPOSE) up

# Restart services (clean restart)
restart:
	@echo "Restarting MCP Gateway Stack..."
	@$(DOCKER_COMPOSE) down --remove-orphans
	@docker container prune -f
	@$(DOCKER_COMPOSE) up

# Clean up Docker resources (safe)
prune:
	@echo "Cleaning up unused Docker resources..."
	@docker container prune -f
	@docker image prune -f
	@docker volume prune -f
	@docker network prune -f
	@echo "Docker cleanup complete"

# Clean up all Docker resources (aggressive)
docker-prune:
	@echo "Aggressive Docker cleanup (removes all unused resources)..."
	@$(DOCKER_COMPOSE) down --remove-orphans
	@docker system prune -a -f --volumes
	@echo "Aggressive Docker cleanup complete"

# Complete Docker reset (nuclear option)
docker-reset:
	@echo "Resetting all project Docker resources..."
	@$(DOCKER_COMPOSE) down -v --remove-orphans --rmi all
	@docker system prune -a -f --volumes
	@echo "Docker reset complete - all project images and volumes removed"

# Database shell
db-shell:
	@echo "Opening PostgreSQL shell..."
	@$(DOCKER_COMPOSE) exec postgres psql -U ${DB_USERNAME} -d ${DB_DATABASE}

# Redis CLI
redis-cli:
	@echo "Opening Redis CLI..."
	@$(DOCKER_COMPOSE) exec redis redis-cli -a ${REDIS_PASSWORD}

# Watch for file changes (requires air installed locally)
watch:
	@echo "Starting local development with hot reload..."
	@air

# Nuclear strike - destroy everything
nuclear:
	@echo "Initiating nuclear strike..."
	@docker stop $$(docker ps -q) 2>/dev/null || true
	@docker rm $$(docker ps -aq) 2>/dev/null || true
	@docker rmi $$(docker images -q) -f 2>/dev/null || true
	@docker volume rm $$(docker volume ls -q) 2>/dev/null || true
	@docker network rm $$(docker network ls -q --filter type=custom) 2>/dev/null || true
	@docker system prune -a --volumes -f 2>/dev/null || true
	@rm -rf tmp/ bin/ dist/
	@rm -f main api apps/backend/api logs/*.log
	@go clean -modcache 2>/dev/null || true
	@go clean -testcache 2>/dev/null || true
	@go clean -cache 2>/dev/null || true
	@echo "Boom! 💥"
