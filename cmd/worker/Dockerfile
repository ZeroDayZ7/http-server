# --- Stage 1: Build binary ---
FROM golang:1.26-alpine AS builder

# Install build dependencies and security updates
RUN apk add --no-cache ca-certificates git tzdata

WORKDIR /src

# Leverage Docker cache for dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source and build static binary
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-w -s" -o /app/worker ./cmd/worker


# --- Stage 2: Final hardened runtime ---
FROM alpine:3.19 AS final

# Standard production tools (SSL certs and Timezones)
RUN apk add --no-cache ca-certificates tzdata

# Create dedicated system user (UID > 10000 for host isolation)
RUN addgroup -S workergroup && \
    adduser -S -u 10002 -g workergroup workeruser

WORKDIR /home/workeruser

# Securely copy binary and set non-root ownership
COPY --from=builder --chown=workeruser:workergroup /app/worker .

# Strip write permissions to ensure immutability (Security Best Practice)
RUN chmod 500 ./worker

# Switch to unprivileged user
USER workeruser

# Signal handling: use JSON syntax to allow graceful shutdowns (SIGTERM)
ENTRYPOINT ["./worker"]