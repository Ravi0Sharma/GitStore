# Multi-stage build for Go backend server
FROM golang:1.24-alpine AS builder

# Install git for go modules that might need it
RUN apk add --no-cache git

# Copy go mod files first for better caching (context is repo root)
# Maintain directory structure: gitClone/ and gitDb/ at same level
COPY gitClone/go.mod /build/gitClone/go.mod
COPY gitDb/go.mod /build/gitDb/go.mod

# Set working directory to gitClone (where the main module is)
WORKDIR /build/gitClone

# Download dependencies
RUN go mod download

# Copy source code (maintaining relative structure)
COPY gitClone/ /build/gitClone/
COPY gitDb/ /build/gitDb/

# Build the server binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o gitserver ./cmd/gitserver

# Final stage: minimal runtime image
FROM alpine:latest

# Install ca-certificates and wget for healthchecks
RUN apk --no-cache add ca-certificates wget

WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/gitClone/gitserver .

# Create data directories
RUN mkdir -p /data/repos /data/db

# Expose port
EXPOSE 8080

# Run the server
CMD ["./gitserver"]

