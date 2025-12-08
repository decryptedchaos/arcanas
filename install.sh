#!/bin/bash

# NAS Dashboard Production Installer
# Downloads prebuilt release and sets up systemd service

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
INSTALL_DIR="/opt/arcanas"
SERVICE_USER="arcanas"
SERVICE_NAME="arcanas"
REPO_OWNER="decryptedchaos"
REPO_NAME="arcanas"
VERSION="latest"

print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if running as root
if [[ $EUID -ne 0 ]]; then
   print_error "This script must be run as root (use sudo)"
   exit 1
fi

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --version)
            VERSION="$2"
            shift 2
            ;;
        --repo)
            REPO_OWNER="$2"
            shift 2
            ;;
        --help)
            echo "Usage: $0 [--version VERSION] [--repo OWNER/REPO]"
            echo "Example: $0 --version v1.0.0 --repo myusername/nas-dashboard"
            exit 0
            ;;
        *)
            print_error "Unknown option: $1"
            exit 1
            ;;
    esac
done

print_status "Starting NAS Dashboard installation..."
print_status "Repository: $REPO_OWNER/$REPO_NAME"
print_status "Version: $VERSION"

# Function to detect OS
detect_os() {
    if [ -f /etc/os-release ]; then
        . /etc/os-release
        OS=$NAME
        VER=$VERSION_ID
    else
        print_error "Cannot detect operating system"
        exit 1
    fi
    print_status "Detected OS: $OS $VER"
}

# Function to install system dependencies
install_dependencies() {
    print_status "Installing system dependencies..."
    
    case $OS in
        "Ubuntu"*|"Debian"*)
            apt update
            apt install -y curl wget tar systemd
            ;;
        "CentOS"*|"Red Hat"*|"Fedora"*)
            if command -v dnf &> /dev/null; then
                dnf install -y curl wget tar systemd
            else
                yum install -y curl wget tar systemd
            fi
            ;;
        "Arch Linux"*)
            pacman -Sy --noconfirm curl wget tar systemd
            ;;
        *)
            print_error "Unsupported OS: $OS"
            exit 1
            ;;
    esac
    
    print_success "System dependencies installed"
}

# Function to download release
download_release() {
    print_status "Downloading NAS Dashboard release..."
    
    cd /tmp
    
    # Get download URL
    if [ "$VERSION" = "latest" ]; then
        DOWNLOAD_URL=$(curl -s "https://api.github.com/repos/$REPO_OWNER/$REPO_NAME/releases/latest" | grep "browser_download_url.*linux-amd64.tar.gz" | cut -d '"' -f 4)
    else
        DOWNLOAD_URL="https://github.com/$REPO_OWNER/$REPO_NAME/releases/download/$VERSION/nas-dashboard-${VERSION#v}-linux-amd64.tar.gz"
    fi
    
    if [ -z "$DOWNLOAD_URL" ]; then
        print_error "Could not find download URL for version $VERSION"
        print_error "Make sure the repository exists and has releases"
        exit 1
    fi
    
    print_status "Downloading from: $DOWNLOAD_URL"
    wget -q --show-progress "$DOWNLOAD_URL" -O nas-dashboard.tar.gz
    
    if [ ! -f nas-dashboard.tar.gz ]; then
        print_error "Download failed"
        exit 1
    fi
    
    print_success "Release downloaded"
}

# Function to create service user
create_service_user() {
    if id "$SERVICE_USER" &>/dev/null; then
        print_status "Service user $SERVICE_USER already exists"
        return
    fi
    
    print_status "Creating service user: $SERVICE_USER"
    useradd -r -s /bin/false -d $INSTALL_DIR $SERVICE_USER
    print_success "Service user created"
}

# Function to install files
install_files() {
    print_status "Installing NAS Dashboard files..."
    
    # Create installation directory
    mkdir -p $INSTALL_DIR
    
    # Extract archive
    cd /tmp
    tar -xzf nas-dashboard.tar.gz
    
    # Copy files to installation directory
    cp -r nas-dashboard/* $INSTALL_DIR/
    
    # Set ownership and permissions
    chown -R $SERVICE_USER:$SERVICE_USER $INSTALL_DIR
    chmod +x $INSTALL_DIR/nas-dashboard
    
    # Cleanup
    rm -rf /tmp/nas-dashboard*
    
    print_success "Files installed to $INSTALL_DIR"
}

# Function to create systemd service
create_systemd_service() {
    print_status "Creating systemd service..."
    
    cat > /etc/systemd/system/${SERVICE_NAME}.service << EOF
[Unit]
Description=NAS Dashboard
After=network.target

[Service]
Type=simple
User=$SERVICE_USER
Group=$SERVICE_USER
WorkingDirectory=$INSTALL_DIR
ExecStart=$INSTALL_DIR/nas-dashboard
Restart=always
RestartSec=5
Environment=API_PORT=4000

# Security settings
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=$INSTALL_DIR

[Install]
WantedBy=multi-user.target
EOF

    # Reload systemd
    systemctl daemon-reload
    
    print_success "Systemd service created"
}

# Function to setup firewall
setup_firewall() {
    print_status "Setting up firewall rules..."
    
    if command -v ufw &> /dev/null; then
        ufw allow 4000/tcp
        print_status "Firewall configured (ufw)"
    elif command -v firewall-cmd &> /dev/null; then
        firewall-cmd --permanent --add-port=4000/tcp
        firewall-cmd --reload
        print_status "Firewall configured (firewalld)"
    else
        print_warning "Could not configure firewall automatically"
    fi
}

# Function to enable and start service
start_service() {
    print_status "Enabling and starting NAS Dashboard service..."
    
    systemctl enable $SERVICE_NAME
    systemctl start $SERVICE_NAME
    
    # Wait a moment and check status
    sleep 2
    if systemctl is-active --quiet $SERVICE_NAME; then
        print_success "NAS Dashboard service is running"
        print_status "Access at: http://$(hostname -I | awk '{print $1}'):4000"
    else
        print_error "Failed to start NAS Dashboard service"
        systemctl status $SERVICE_NAME
        exit 1
    fi
}

# Function to show post-install info
show_info() {
    echo ""
    print_success "NAS Dashboard installation completed successfully!"
    echo ""
    echo "Service Management:"
    echo "  Start:   sudo systemctl start $SERVICE_NAME"
    echo "  Stop:    sudo systemctl stop $SERVICE_NAME"
    echo "  Restart: sudo systemctl restart $SERVICE_NAME"
    echo "  Status:  sudo systemctl status $SERVICE_NAME"
    echo "  Logs:    sudo journalctl -u $SERVICE_NAME -f"
    echo ""
    echo "Access NAS Dashboard at:"
    echo "  http://$(hostname -I | awk '{print $1}'):4000"
    echo ""
    echo "Installation directory: $INSTALL_DIR"
    echo "Installed version: $VERSION"
    echo ""
}

# Main installation flow
main() {
    detect_os
    install_dependencies
    download_release
    create_service_user
    install_files
    create_systemd_service
    setup_firewall
    start_service
    show_info
}

# Run main function
main "$@"
