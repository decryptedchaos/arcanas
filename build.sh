#!/bin/bash

# Arcanas Build Script
# Creates a single binary with embedded frontend

set -e

echo "🔨 Building Arcanas..."

# Check if we're in the right directory
if [ ! -f "backend/go.mod" ] || [ ! -f "frontend/package.json" ]; then
    echo "❌ Error: Run this script from the arcanas root directory"
    exit 1
fi

# Build frontend
echo "📦 Building frontend..."
cd frontend
npm run build

# Check if frontend build succeeded
if [ ! -d "build" ]; then
    echo "❌ Error: Frontend build failed - no build directory created"
    exit 1
fi

cd ..

# Build Go binary with embedded frontend
echo "🔧 Building Go binary..."
cd backend

# Set build info
BUILD_TIME=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
GIT_COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
VERSION=${1:-"1.0.0"}

# Build for current platform
go build -ldflags "-X main.Version=$VERSION -X main.BuildTime=$BUILD_TIME -X main.GitCommit=$GIT_COMMIT" \
    -o ../arcanas cmd/server/main.go

cd ..

echo "✅ Build complete!"
echo "📁 Binary: ./arcanas"
echo ""
echo "🚀 To run:"
echo "  ./arcanas"
echo ""
echo "🌐 Web UI will be available at: http://localhost:4000"
echo "📡 API will be available at: http://localhost:4000/api"
echo ""
echo "🔧 Environment variables:"
echo "  API_PORT=4000    # API server port (default: 4000)"

