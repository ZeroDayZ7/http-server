# 1. builder
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Debug: pokaż co jest w systemie
RUN pwd && ls -la

# Kopiowanie modułów
COPY go.mod go.sum ./
RUN ls -la && go mod download

# Kopiowanie całego repo
COPY . .

# 🔥 DEBUG: co faktycznie przyszło do kontenera
RUN echo "===== ROOT DIR =====" && ls -la
RUN echo "===== FIND ALL FILES =====" && find . -type f
RUN echo "===== GO FILES =====" && find . -name "*.go"

# Debug: pokaż strukturę
RUN echo "===== TREE (if available) =====" && (which tree && tree || echo "tree not installed")

# Build
RUN echo "===== BUILD START =====" && \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -v -o app . && \
    echo "===== BUILD SUCCESS =====" && ls -la

# 2. runtime
FROM alpine:latest

WORKDIR /app

# Debug runtime
RUN ls -la

COPY --from=builder /app/app .

RUN echo "===== RUNTIME FILES =====" && ls -la

CMD ["./app"]