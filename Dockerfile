FROM golang:1.24-alpine

# Install SQLite dependencies (needed for go-sqlite3 CGO bindings)
RUN apk add --no-cache gcc musl-dev sqlite-dev

WORKDIR /app

# Copy go mod files first (layer caching optimization)
COPY src/go.mod src/go.sum ./
RUN go mod download

# Copy source code
COPY src/ .

# Build with CGO enabled for SQLite
RUN CGO_ENABLED=1 go build -o main ./backend/app.go

# Expose application port
EXPOSE 8080

# Set default database path (can be overridden)
ENV DATABASE_URL=/app/data/oggole.db

# Run the application
CMD ["./main"]

