#!/bin/bash

# NAS Dashboard Development Setup
# Sets up development environment with all dependencies

set -e

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m'

print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

# Check if running as root for system deps
if [[ $EUID -ne 0 ]]; then
   echo "This script installs system dependencies. Run with sudo for system deps, or without for project setup only."
   SYSTEM_DEPS=false
else
   SYSTEM_DEPS=true
fi

# Install system dependencies if root
if [ "$SYSTEM_DEPS" = true ]; then
    print_status "Installing system dependencies..."
    
    # Detect OS
    if [ -f /etc/os-release ]; then
        . /etc/os-release
        OS=$NAME
    else
        echo "Cannot detect OS"
        exit 1
    fi
    
    case $OS in
        "Ubuntu"*|"Debian"*)
            apt update
            apt install -y curl git build-essential
            ;;
        "CentOS"*|"Red Hat"*|"Fedora"*)
            if command -v dnf &> /dev/null; then
                dnf install -y curl git gcc
            else
                yum install -y curl git gcc
            fi
            ;;
        "Arch Linux"*)
            pacman -Sy --noconfirm curl git base-devel
            ;;
    esac
    
    print_success "System dependencies installed"
fi

# Setup Go (if not root, check if already installed)
if ! command -v go &> /dev/null; then
    print_status "Installing Go..."
    cd /tmp
    wget -q https://go.dev/dl/go1.21.5.linux-amd64.tar.gz
    if [ "$SYSTEM_DEPS" = true ]; then
        tar -C /usr/local -xzf go1.21.5.linux-amd64.tar.gz
        echo 'export PATH=$PATH:/usr/local/go/bin' > /etc/profile.d/go.sh
    else
        tar -C $HOME -xzf go1.21.5.linux-amd64.tar.gz
        echo 'export PATH=$PATH:$HOME/go/bin' >> $HOME/.bashrc
    fi
    rm -f go1.21.5.linux-amd64.tar.gz
    print_success "Go installed"
else
    print_status "Go is already installed"
fi

# Setup Node.js (if not root, check if already installed)
if ! command -v node &> /dev/null; then
    print_status "Installing Node.js..."
    cd /tmp
    wget -q https://nodejs.org/dist/v18/node-v18-linux-x64.tar.xz
    if [ "$SYSTEM_DEPS" = true ]; then
        tar -C /usr/local -xf node-v18-linux-x64.tar.xz
        echo 'export PATH=$PATH:/usr/local/node/bin' > /etc/profile.d/node.sh
    else
        tar -C $HOME -xf node-v18-linux-x64.tar.xz
        echo 'export PATH=$PATH:$HOME/node/bin' >> $HOME/.bashrc
    fi
    rm -f node-v18-linux-x64.tar.xz
    print_success "Node.js installed"
else
    print_status "Node.js is already installed"
fi

# Setup project dependencies
print_status "Setting up project dependencies..."

# Frontend dependencies
cd frontend
print_status "Installing frontend dependencies..."
npm install

# Backend dependencies
cd ../backend
print_status "Installing backend dependencies..."
go mod download

print_success "Development environment setup complete!"
echo ""
echo "To run NAS Dashboard in development mode:"
echo "  cd backend && go run cmd/server/main.go"
echo ""
echo "Frontend will auto-build when DEV_MODE=true"
