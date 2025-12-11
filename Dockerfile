# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /build

# Copy go mod files first (layer caching optimization)
COPY src/go.mod src/go.sum ./
RUN go mod download

# Copy source code
COPY src/ .

# Build static binary (no CGO needed for PostgreSQL)
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-w -s" \
    -o oggole \
    ./backend

# Runtime stage
FROM alpine:3.21

# Install runtime dependencies only
RUN apk add --no-cache ca-certificates=20250911-r0 wget=1.25.0-r0 && \
    addgroup -g 1000 app && \
    adduser -D -u 1000 -G app app

WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /build/oggole .

# Copy static files and templates
COPY --from=builder /build/static ./static
COPY --from=builder /build/templates ./templates

# Security: run as non-root user
USER app

# Expose application port
EXPOSE 8080

# Run the application
CMD ["./oggole"]
