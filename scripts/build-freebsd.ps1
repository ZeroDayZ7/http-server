# Setting up environment for FreeBSD build
$env:CGO_ENABLED = "0"
$env:GOOS = "freebsd"
$env:GOARCH = "amd64"

# --- 1. Kompilacja API ---
Write-Host "Building API..." -ForegroundColor Cyan
go build -ldflags="-s -w" -o "./bin/api_server" "./cmd/api"

# --- 2. Kompilacja WORKERA ---
Write-Host "Building Worker..." -ForegroundColor Cyan
go build -ldflags="-s -w" -o "./bin/worker_service" "./cmd/worker"

# --- OPCJONALNIE: Garble ---
# garble -literals -seed=random build -trimpath -ldflags="-s -w" -o "./bin/api_server" ./cmd/api
# garble -literals -seed=random build -trimpath -ldflags="-s -w" -o "./bin/worker_service" ./cmd/worker

# --- 3. Generowanie skrótów SHA256 ---
Write-Host "Generating hashes..." -ForegroundColor Yellow
Get-FileHash "./bin/api_server" -Algorithm SHA256 | Out-File "./bin/api_server.sha256"
Get-FileHash "./bin/worker_service" -Algorithm SHA256 | Out-File "./bin/worker_service.sha256"

# --- Post-compilation message ---
Write-Host "`nBuild complete!" -ForegroundColor Green
Write-Host "Binary 1: ./bin/api_server"
Write-Host "Binary 2: ./bin/worker_service"