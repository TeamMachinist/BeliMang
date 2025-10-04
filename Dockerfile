# Stage 1: PostgreSQL with H3 extension
FROM postgres:18 AS postgres-h3

RUN apt-get update && \
    apt-get install -y --no-install-recommends \
    ca-certificates \
    postgresql-18-h3 && \
    rm -rf /var/lib/apt/lists/*

# Stage 2: Go builder with CGO enabled
FROM golang:1.25-bookworm AS builder

# Install build dependencies
RUN apt-get update && apt-get install -y --no-install-recommends \
    git \
    build-essential \
    ca-certificates && \
    rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Copy dependency files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download && go mod verify

# Copy source code
COPY . .

# Build with CGO enabled and optimizations
ENV CGO_ENABLED=1 \
    GOOS=linux \
    GOARCH=amd64

# Build the application from cmd/main.go
RUN go build -v \
    -ldflags="-s -w" \
    -o /app/bin/server \
    ./cmd

# Stage 3: Production runtime
FROM debian:bookworm-slim AS production

# Install runtime dependencies
RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    tzdata && \
    rm -rf /var/lib/apt/lists/*

# Create non-root user
RUN groupadd -r appuser && useradd -r -g appuser appuser

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/bin/server /app/server
COPY --from=builder --chown=appuser:appuser /app/migrations /app/migrations

# Set ownership
RUN chown -R appuser:appuser /app

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=40s --retries=3 \
    CMD ["/app/server", "health"] || exit 1

# Run the application
CMD ["/app/server"]