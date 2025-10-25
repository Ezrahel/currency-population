# Build stage
FROM golang:1.21-alpine AS builder

# Install required packages
RUN apk add --no-cache gcc musl-dev

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=1 GOOS=linux go build -o /app/main ./cmd/api/main.go

# Final stage
FROM alpine:latest

# Install required packages for runtime
RUN apk add --no-cache ca-certificates tzdata freetype-dev

# Create non-root user
RUN adduser -D appuser

# Create necessary directories
RUN mkdir -p /app/cache && \
    chown -R appuser:appuser /app

# Set working directory
WORKDIR /app

# Copy the binary from builder
COPY --from=builder /app/main .

# Copy .env file
COPY .env .

# Set user
USER appuser

# Expose port
EXPOSE 8080

# Run the application
CMD ["./main"]