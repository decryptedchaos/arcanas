#!/bin/bash

# Development script with Air hot-reloading
# This script ensures Air is available and starts the backend with hot-reload

set -e

echo "🔥 Starting Arcanas backend with Air hot-reload..."

# Check if Air is installed
AIR_PATH=$(which air 2>/dev/null || echo "")

if [ -z "$AIR_PATH" ]; then
    # Try Go bin directory
    GOPATH=$(go env GOPATH)
    if [ -f "$GOPATH/bin/air" ]; then
        AIR_PATH="$GOPATH/bin/air"
        echo "✓ Found Air at: $AIR_PATH"
    else
        echo "❌ Air not found. Installing..."
        go install github.com/air-verse/air@latest
        AIR_PATH="$GOPATH/bin/air"
    fi
fi

# Run Air
echo "🚀 Starting backend with hot-reload..."
echo "📝 Backend will auto-restart on file changes"
echo ""

cd backend
exec "$AIR_PATH"
