# Synced with go.mod version
FROM golang:1.22-alpine AS builder

# Install build dependencies required for go-sqlite3 (CGO)
RUN apk add --no-cache build-base

WORKDIR /app

# Copy dependency files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
# CGO_ENABLED=1 is required for go-sqlite3
RUN CGO_ENABLED=1 GOOS=linux go build -ldflags="-w -s" -o wordflow cmd/wordflow/main.go

# Final stage
FROM alpine:latest

WORKDIR /app

# Install runtime dependencies (ca-certificates for HTTPS, tzdata for timezone)
RUN apk add --no-cache ca-certificates tzdata

# Copy the binary from builder
COPY --from=builder /app/wordflow .

# Expose the default port
EXPOSE 8080

# Set the entrypoint
ENTRYPOINT ["./wordflow"]

# Default command
CMD ["server"]
