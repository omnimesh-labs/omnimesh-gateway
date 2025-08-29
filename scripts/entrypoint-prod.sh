#!/bin/bash

set -e

log() {
    echo "[prod] $@"
}

log "Starting production environment..."

# Wait for PostgreSQL
log "Waiting for PostgreSQL..."
/scripts/wait-for-it.sh ${DB_HOST}:${DB_PORT} -t 60

# Run migrations using binary
log "Running migrations..."
/app/migrate up || { log "Migration failed"; exit 1; }
log "Migrations completed"

# Start application
log "Starting application..."
exec /app/main
