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

# Run complete setup (admin user + org + dummy data)
if [ "${SKIP_SETUP}" != "true" ]; then
    log "Running complete setup..."
    cd /app && go run apps/backend/cmd/setup/main.go all || log "Setup completed or data already exists"
fi

# Start application
if [ "${MODE}" = "development" ]; then
    exec air
else
    exec /app/main
fi