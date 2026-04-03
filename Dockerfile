# 1. builder
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Debug
RUN pwd && ls -la

# Moduły
COPY go.mod go.sum ./
RUN ls -la && go mod download

# Kod
COPY . .

# Debug – co faktycznie jest w kontenerze
RUN echo "===== ROOT DIR =====" && ls -la
RUN echo "===== FIND ALL FILES =====" && find . -type f
RUN echo "===== GO FILES =====" && find . -name "*.go"

# Debug struktury
RUN echo "===== TREE =====" && (which tree && tree || echo "tree not installed")

# Build API (KLUCZOWA ZMIANA)
RUN echo "===== BUILD START =====" && \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -v -o app ./cmd/api && \
    echo "===== BUILD SUCCESS =====" && ls -la

# 2. runtime
FROM alpine:latest

WORKDIR /app

RUN ls -la

COPY --from=builder /app/app .

RUN echo "===== RUNTIME FILES =====" && ls -la

CMD ["./app"]