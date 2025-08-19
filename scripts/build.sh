#!/bin/bash

# Build script for MCP Gateway

set -e

# Configuration
APP_NAME="mcp-gateway"
VERSION=${VERSION:-"1.0.0"}
BUILD_TIME=$(date -u +%Y-%m-%dT%H:%M:%SZ)
GIT_COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_USER=$(whoami)

# Go build flags
LDFLAGS="-X main.version=${VERSION} -X main.buildTime=${BUILD_TIME} -X main.gitCommit=${GIT_COMMIT} -X main.buildUser=${BUILD_USER}"

echo "Building ${APP_NAME} version ${VERSION}..."

# Create bin directory if it doesn't exist
mkdir -p bin

# Build main API server
echo "Building API server..."
go build -ldflags "${LDFLAGS}" -o bin/${APP_NAME}-api ./cmd/api

# Build migration tool
echo "Building migration tool..."
go build -ldflags "${LDFLAGS}" -o bin/${APP_NAME}-migrate ./cmd/migrate

# Build worker
echo "Building worker..."
go build -ldflags "${LDFLAGS}" -o bin/${APP_NAME}-worker ./cmd/worker

echo "Build completed successfully!"
echo "Binaries:"
echo "  - bin/${APP_NAME}-api"
echo "  - bin/${APP_NAME}-migrate"
echo "  - bin/${APP_NAME}-worker"
