# Use official Go image with Alpine Linux
FROM golang:1.24-alpine

# Set working directory inside container
WORKDIR /app

# Copy Go module files first (for layer caching)
COPY src/go.mod src/go.sum ./
RUN go mod download

# Copy all source code
COPY src/ .

# Build the application
RUN go build -o oggole ./backend

# Expose the application port
EXPOSE 8080

# Run the application
CMD ["./oggole"]
