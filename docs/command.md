go test -v -run TestInteractionWorker_DBRecovery_NoDuplicates ./internal/worker/...
docker compose up -d grafana

# Sprawdź swojego Workera (Go)
curl -v http://localhost:9001/health

# Sprawdź swoje API
curl -v http://localhost:9000/health