FROM golang:1.26-alpine AS builder

RUN apk add --no-cache ca-certificates git tzdata

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -trimpath -ldflags="-w -s" -o /app/server ./cmd/api

FROM alpine:3.19.1 AS final

LABEL org.opencontainers.image.source="https://github.com/ZeroDayZ7/http-server"
LABEL org.opencontainers.image.description="Hardened Go API"

RUN apk update && \
    apk add --no-cache ca-certificates tzdata && \
    rm -rf /var/cache/apk/*

RUN addgroup -S appgroup && \
    adduser -S -u 10001 -g appgroup appuser

WORKDIR /home/appuser

COPY --from=builder --chown=appuser:appgroup /app/server .
COPY --from=builder --chown=appuser:appgroup /src/migrations ./migrations

RUN chmod 500 ./server && \
    chmod 500 ./migrations && \
    chmod 400 ./migrations/*

USER appuser

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

ENTRYPOINT ["./server"]