# Build stage
FROM golang:1.21-alpine AS builder

# Set working directory
WORKDIR /app

# Install git (required for some Go modules)
RUN apk add --no-cache git

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

# Create a non-root user
RUN adduser -D -s /bin/sh appuser

# Set working directory
WORKDIR /root/

# Copy the binary from builder stage
COPY --from=builder /app/main .

# Change ownership to appuser
RUN chown appuser:appuser main

# Switch to non-root user
USER appuser

# Expose port 5465
EXPOSE 5465

# Command to run the application
CMD ["./main"]

