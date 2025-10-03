# Build stage
FROM golang:1.24-alpine AS builder

# Install SQLite dependencies (needed for go-sqlite3 CGO bindings)
RUN apk add --no-cache gcc musl-dev sqlite-dev

WORKDIR /build

# Copy go mod files first (layer caching optimization)
COPY src/go.mod src/go.sum ./
RUN go mod download

# Copy source code
COPY src/ .

# Build with CGO enabled for SQLite, with size optimizations
RUN CGO_ENABLED=1 GOOS=linux go build \
    -ldflags="-w -s" \
    -o oggole \
    ./backend/app.go

# Runtime stage
FROM alpine:3.21

# Install runtime dependencies only
RUN apk add --no-cache sqlite-libs ca-certificates && \
    addgroup -g 1000 app && \
    adduser -D -u 1000 -G app app

WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /build/oggole .

# Security: run as non-root user
USER app

# Expose application port
EXPOSE 8080

# Set default database path (can be overridden)
ENV DATABASE_URL=/app/data/oggole.db

# Run the application
CMD ["./oggole"]
