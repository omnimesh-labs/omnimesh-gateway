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

# Build migrate binary if needed
if [ "${MODE}" = "development" ] || [ ! -f /app/migrate ]; then
    log "Building migrate binary..."
    go build -o /app/migrate apps/backend/cmd/migrate/main.go
fi

# Run migrations
log "Running migrations..."
/app/migrate up
log "Migrations completed"

# Setup admin user
if [ "${SKIP_ADMIN_SETUP}" != "true" ]; then
    /app/migrate setup-admin || log "Admin user already exists"
fi

# Start application
if [ "${MODE}" = "development" ]; then
    exec air
else
    exec /app/main
fi