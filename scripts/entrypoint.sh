#!/bin/bash

set -e

log() {
    echo "[entrypoint] $@"
}

MODE="${MODE:-development}"
log "Starting in ${MODE} mode"

# Wait for PostgreSQL
log "Waiting for PostgreSQL..."
/scripts/wait-for-it.sh ${DB_HOST}:${DB_PORT} -t 60

# Run migrations
log "Running migrations..."
/app/migrate up || { log "Migration failed"; exit 1; }
log "Migrations completed"

# Run setup only if explicitly enabled
[ "${MODE}" = "development" ] && {
    log "Running setup..."
    /app/setup all 2>/dev/null || log "Setup completed or skipped"
}

# Start application
[ "${MODE}" = "development" ] && exec air || exec /app/main
