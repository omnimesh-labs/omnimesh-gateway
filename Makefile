# Check if .env exists and include it
ifeq (,$(wildcard ./.env))
    $(error .env file not found. Please create one.)
endif

include .env
export

# Simple Makefile for a Go project

# Build the application
all: build test

build:
	@echo "Building..."
	
	
	@go build -o main apps/backend/cmd/api/main.go

# Run the application
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

# Integrations Tests for the application
itest:
	@echo "Running integration tests..."
	@go test ./apps/backend/internal/database -v

# Clean the binary
clean:
	@echo "Cleaning..."
	@rm -f main

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

.PHONY: all build run test clean watch docker-run docker-down itest migrate migrate-down migrate-status migrate-create