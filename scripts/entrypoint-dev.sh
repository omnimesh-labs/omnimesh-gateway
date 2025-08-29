#!/bin/bash

set -e

log() {
    echo "[dev] $@"
}

log "Starting development environment..."

# Wait for PostgreSQL
log "Waiting for PostgreSQL..."
/scripts/wait-for-it.sh ${DB_HOST}:${DB_PORT} -t 60

# Run migrations using go run
log "Running migrations..."
go run apps/backend/cmd/migrate/main.go up || { log "Migration failed"; exit 1; }
log "Migrations completed"

# Run setup using go run
log "Running setup..."
go run apps/backend/cmd/setup/main.go all || { log "Setup failed"; exit 1; }
log "Setup completed"

# Start application with hot reload
log "Starting with hot reload..."
exec air
