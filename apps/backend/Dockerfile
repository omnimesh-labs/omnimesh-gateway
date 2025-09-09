# Development stage with air for hot reload
FROM golang:1.23-alpine AS development

WORKDIR /app

# Install dependencies including air, bash, make, Python, Node.js, uvx, and npx
RUN apk add --no-cache git gcc musl-dev bash make python3 python3-dev py3-pip curl nodejs npm && \
    go install github.com/air-verse/air@v1.61.1 && \
    curl -LsSf https://astral.sh/uv/install.sh | sh && \
    ln -sf /root/.local/bin/uv /usr/local/bin/uv && \
    ln -sf /root/.local/bin/uvx /usr/local/bin/uvx

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Copy scripts
COPY scripts/wait-for-it.sh /scripts/
COPY scripts/entrypoint-dev.sh /scripts/
RUN chmod +x /scripts/*.sh

# Copy air config
COPY .air.toml ./

# Environment for development mode
ENV MODE=development
ENV PATH="$PATH:/root/.local/bin"

# Expose port
EXPOSE 8080

# Use development entrypoint script
ENTRYPOINT ["/scripts/entrypoint-dev.sh"]

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

# Build main application, migrate tool, and setup tool
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o main apps/backend/cmd/api/main.go && \
    CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o migrate apps/backend/cmd/migrate/main.go && \
    CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o setup apps/backend/cmd/setup/main.go

# Production stage
FROM alpine:latest AS production

RUN apk --no-cache add ca-certificates tzdata bash python3 python3-dev py3-pip curl nodejs npm && \
    curl -LsSf https://astral.sh/uv/install.sh | sh && \
    ln -sf /root/.local/bin/uv /usr/local/bin/uv && \
    ln -sf /root/.local/bin/uvx /usr/local/bin/uvx

WORKDIR /app

# Copy the binaries from builder
COPY --from=builder /app/main .
COPY --from=builder /app/migrate .
COPY --from=builder /app/setup .
COPY --from=builder /app/apps/backend/configs ./apps/backend/configs
COPY --from=builder /app/apps/backend/migrations ./apps/backend/migrations

# Copy scripts
COPY scripts/wait-for-it.sh /scripts/
COPY scripts/entrypoint-prod.sh /scripts/
RUN chmod +x /scripts/*.sh

# Environment for production mode
ENV MODE=production
ENV PATH="$PATH:/root/.local/bin"

# Expose port
EXPOSE 8080

# Use production entrypoint script
ENTRYPOINT ["/scripts/entrypoint-prod.sh"]
