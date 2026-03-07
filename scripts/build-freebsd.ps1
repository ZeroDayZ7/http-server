# Setting up environment for FreeBSD build
$env:CGO_ENABLED = "0"
$env:GOOS = "freebsd"
$env:GOARCH = "amd64"

# Compile the Go project
go build -o "./bin/server" "./cmd"

# Final
# garble -literals -seed=random build `
# -trimpath `
# -ldflags="-s -w" `
# -o "./bin/server" `
# ./cmd

# Generate SHA256 hash of binary
Get-FileHash "./bin/server" -Algorithm SHA256 |
Out-File "./bin/server.sha256"


# Post-compilation message
Write-Host "Build complete: ./bin/server"
