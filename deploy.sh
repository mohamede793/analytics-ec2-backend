#!/bin/bash
set -e

echo "🚀 Deploying Go API..."

# Pull latest code
git pull origin main

# Install Go if not already installed
if ! command -v go &> /dev/null; then
    echo "Installing Go..."
    wget -q https://go.dev/dl/go1.21.5.linux-amd64.tar.gz
    sudo tar -C /usr/local -xzf go1.21.5.linux-amd64.tar.gz
    echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
    export PATH=$PATH:/usr/local/go/bin
    rm go1.21.5.linux-amd64.tar.gz
else
    export PATH=$PATH:/usr/local/go/bin
fi

# Install dependencies
go mod tidy

# Build the application
go build -o api-server main.go

# Stop existing server
pkill -f api-server || true
sleep 2

# Start new server
nohup ./api-server > app.log 2>&1 &

echo "✅ Deployment complete!"
echo "Server PID: $(pgrep -f api-server)"
echo "Recent logs:"
tail -n 5 app.log