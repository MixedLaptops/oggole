# Use official Go image with Alpine Linux
FROM golang:1.24-alpine

# Create non-root user for security (privilege separation)
RUN adduser -D appuser

# Set working directory inside container
WORKDIR /app

# Copy Go module files first (for layer caching)
COPY src/go.mod src/go.sum ./
RUN go mod download

# Copy all source code
COPY src/ .

# Build the application
RUN go build -o oggole ./backend

# Change ownership so non-root user can execute
RUN chown -R appuser:appuser /app

# Switch to non-root user
USER appuser

# Expose the application port
EXPOSE 8080

# Run the application
CMD ["./oggole"]
