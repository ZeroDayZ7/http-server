FROM golang:1.26-alpine AS builder

RUN apk add --no-cache ca-certificates git tzdata
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .

# Dodano -trimpath dla bezpieczeństwa logów
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -trimpath -ldflags="-w -s" -o /app/worker ./cmd/worker

FROM alpine:3.19 AS final

# Dodano LABEL dla porządku w GitHubie
LABEL org.opencontainers.image.source="https://github.com/ZeroDayZ7/http-server"
LABEL org.opencontainers.image.description="Hardened Go Worker"

# Dodano apk update i czyszczenie cache
RUN apk update && \
    apk add --no-cache ca-certificates tzdata && \
    rm -rf /var/cache/apk/*

RUN addgroup -S workergroup && \
    adduser -S -u 10002 -g workergroup workeruser

WORKDIR /home/workeruser

COPY --from=builder --chown=workeruser:workergroup /app/worker .

RUN chmod 500 ./worker

USER workeruser

ENTRYPOINT ["./worker"]
