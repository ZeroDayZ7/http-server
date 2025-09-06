# Setting up environment for FreeBSD build
$env:GOOS = "freebsd"
$env:GOARCH = "amd64"

# Compile the Go project
go build -o "../bin/server" "../cmd"

# Post-compilation message
Write-Host "Build complete: ../bin/server"

# Run the compiled server (uncomment to execute)
# ./bin/server       # Linux/macOS
# .\bin\server.exe   # Windows

# Rebuild script (uncomment to run)
# .\build-freebsd.ps1
