# --- ETAP 1: GOLANG BUILDER ---
FROM golang:1.26-alpine AS builder

# Dodajemy narzędzia niezbędne do kompilacji i bezpieczeństwa
RUN apk add --no-cache ca-certificates git tzdata

WORKDIR /src

# Optymalizacja cache zależności
COPY go.mod go.sum ./
RUN go mod download

# Kopiujemy kod źródłowy
COPY . .

# [PRO TIP] Budowanie z flagami, które czynią binarkę jeszcze bezpieczniejszą:
# -trimpath: usuwa ścieżki z Twojego komputera z binarki (bezpieczeństwo logów)
# -ldflags="-w -s": usuwa symbole debugowania
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -trimpath -ldflags="-w -s" -o /app/server ./cmd/api


# --- ETAP 2: FINAL RUNTIME ---
FROM alpine:3.19.1 AS final

# Etykiety zgodne ze standardem OCI (Google/Amazon to kochają)
LABEL org.opencontainers.image.source="https://github.com/twoj-user/twoj-repo"
LABEL org.opencontainers.image.description="Hardened Go API"

# Instalacja tylko niezbędnych paczek i aktualizacja poprawek bezpieczeństwa
RUN apk update && \
    apk add --no-cache ca-certificates tzdata && \
    rm -rf /var/cache/apk/*

# Tworzymy użytkownika (UID 10001 jest poza standardowym zakresem systemowym)
RUN addgroup -S appgroup && \
    adduser -S -u 10001 -g appgroup appuser

WORKDIR /home/appuser

# Kopiujemy binarkę i migracje z poprawnym właścicielem w jednej warstwie
COPY --from=builder --chown=appuser:appgroup /app/server .
COPY --from=builder --chown=appuser:appgroup /src/migrations ./migrations

# [PRO TIP] Usuwamy powłoki (shell), jeśli nie są potrzebne do healthchecka
# Ale zostawiamy wget dla Twojego HEALTHCHECKa poniżej.

# Zabezpieczenie plików (Read-Only)
RUN chmod 500 ./server && \
    chmod -R 400 ./migrations

USER appuser

EXPOSE 8080

# Ulepszony Healthcheck - upewnij się, że Twoja apka Go ma endpoint /health!
HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Uruchomienie jako proces nr 1
ENTRYPOINT ["./server"]