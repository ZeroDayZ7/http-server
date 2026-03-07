# Setting up environment for FreeBSD build
$env:CGO_ENABLED = "0"
$env:GOOS = "freebsd"
$env:GOARCH = "amd64"

# Compile the Go project
go build -o "./bin/server" "./cmd"

# Post-compilation message
Write-Host "Build complete: ./bin/server"
