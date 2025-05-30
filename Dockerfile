# Build stage
FROM golang:1.24.3-alpine AS builder

# Install git and ca-certificates (needed for fetching dependencies and HTTPS)
RUN apk add --no-cache git ca-certificates

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN GOOS=linux go build -o /app/main .

# Runtime stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

# Create app directory
WORKDIR /app

# Copy the binary from builder stage
COPY --from=builder /app/main .

# Create directory for database
RUN mkdir -p /app/youtubecurator.db

# Create directory for feed mocks (for debug mode)
RUN mkdir -p /app/feed_mocks

COPY build/init.sh /app/init.sh

# Expose any ports if needed (not required for this MVP but good practice)
# EXPOSE 8080

# Run the application
CMD ["sh", "/app/init.sh"] 