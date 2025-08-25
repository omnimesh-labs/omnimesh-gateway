# Development stage with air for hot reload
FROM golang:1.23-alpine AS development

WORKDIR /app

# Install dependencies including air, bash, and make
RUN apk add --no-cache git gcc musl-dev bash make && \
    go install github.com/air-verse/air@v1.61.1

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code for migrate binary
COPY . .

# Build the migrate binary
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o migrate apps/backend/cmd/migrate/main.go

# Copy scripts
COPY scripts/wait-for-it.sh /scripts/
COPY scripts/entrypoint.sh /scripts/
RUN chmod +x /scripts/*.sh

# Copy air config
COPY .air.toml ./

# Environment for development mode
ENV MODE=development

# Expose port
EXPOSE 8080

# Use entrypoint script
ENTRYPOINT ["/scripts/entrypoint.sh"]

# Build stage for production
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git gcc musl-dev

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build both the main application and migrate tool
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o main apps/backend/cmd/api/main.go && \
    CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o migrate apps/backend/cmd/migrate/main.go

# Production stage
FROM alpine:latest AS production

RUN apk --no-cache add ca-certificates tzdata bash

WORKDIR /app

# Copy the binaries from builder
COPY --from=builder /app/main .
COPY --from=builder /app/migrate .
COPY --from=builder /app/apps/backend/configs ./apps/backend/configs
COPY --from=builder /app/apps/backend/migrations ./apps/backend/migrations

# Copy scripts
COPY scripts/wait-for-it.sh /scripts/
COPY scripts/entrypoint.sh /scripts/
RUN chmod +x /scripts/*.sh

# Environment for production mode
ENV MODE=production

# Expose port
EXPOSE 8080

# Use entrypoint script
ENTRYPOINT ["/scripts/entrypoint.sh"]