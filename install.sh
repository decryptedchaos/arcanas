#!/bin/bash

# Arcanas Production Installer
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
# Force VERSION to default value first
VERSION="latest"

while [[ $# -gt 0 ]]; do
    case $1 in
        --version)
            if [[ -n "$2" && "$2" != --* ]]; then
                VERSION="$2"
                shift 2
            else
                print_error "Missing version number after --version"
                exit 1
            fi
            ;;
        --repo)
            if [[ -n "$2" && "$2" != --* ]]; then
                REPO_OWNER="$2"
                shift 2
            else
                print_error "Missing repository after --repo"
                exit 1
            fi
            ;;
        --help)
            echo "Usage: $0 [--version VERSION] [--repo OWNER/REPO]"
            echo "Example: $0 --version v1.0.0 --repo myusername/arcanas"
            exit 0
            ;;
        *)
            print_error "Unknown option: $1"
            echo "Use --help for usage information"
            exit 1
            ;;
    esac
done

# Debug output
echo "DEBUG: VERSION='$VERSION'"
echo "DEBUG: REPO_OWNER='$REPO_OWNER'"
echo "DEBUG: REPO_NAME='$REPO_NAME'"

print_status "Starting Arcanas installation..."
print_status "Repository: $REPO_OWNER/$REPO_NAME"
print_status "Version: $VERSION"

# Validate version format
if [[ "$VERSION" != "latest" && ! "$VERSION" =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
    print_error "Invalid version format: $VERSION"
    print_error "Version must be 'latest' or in format v1.0.0"
    exit 1
fi

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
            apt install -y curl wget tar systemd sudo \
                samba targetcli-fb nfs-kernel-server \
                mergerfs lvm2 mdadm
            ;;
        "CentOS"*|"Red Hat"*|"Fedora"*)
            if command -v dnf &> /dev/null; then
                dnf install -y curl wget tar systemd sudo \
                    samba targetcli nfs-utils \
                    mergerfs lvm2 mdadm
            else
                yum install -y curl wget tar systemd sudo \
                    samba targetcli nfs-utils \
                    mergerfs lvm2 mdadm
            fi
            ;;
        "Arch Linux"*)
            pacman -Sy --noconfirm curl wget tar systemd sudo \
                samba targetcli-fb nfs-utils \
                mergerfs lvm2 mdadm
            ;;
        *)
            print_error "Unsupported OS: $OS"
            exit 1
            ;;
    esac
    
    print_success "System dependencies installed"
}

# Function to setup storage sudoers configuration
setup_storage_sudoers() {
    print_status "Setting up storage sudoers configuration..."
    
    # Ensure sudoers.d directory exists
    mkdir -p /etc/sudoers.d
    
    # Create sudoers file for storage operations
    cat > /etc/sudoers.d/arcanas-storage << EOF
# Arcanas storage operations sudoers configuration
# Allows the arcanas user to run specific storage commands without password

Cmnd_Alias ARCANAS_STORAGE = /bin/mkdir, /bin/mount, /bin/umount, /usr/sbin/vgcreate, /usr/sbin/lvcreate, /sbin/mkfs, /usr/bin/mergerfs, /bin/sh, /usr/bin/sed, /bin/rmdir, /usr/sbin/vgremove, /usr/sbin/lvremove, /usr/sbin/chown, /usr/sbin/mdadm, /usr/bin/true

arcanas ALL=(ALL) NOPASSWD: ARCANAS_STORAGE
EOF
    
    # Set proper permissions
    chmod 440 /etc/sudoers.d/arcanas-storage
    
    print_success "Storage sudoers configuration completed"
}

# Function to download release
download_release() {
    print_status "Downloading Arcanas release..."
    
    # FORCE VERSION TO LATEST TO OVERRIDE ANY WEIRDNESS
    VERSION="latest"
    echo "DEBUG: Forced VERSION to '$VERSION'"
    
    cd /tmp
    
    # Get download URL
    if [ "$VERSION" = "latest" ]; then
        # Get latest release info from GitHub API
        LATEST_RELEASE_INFO=$(curl -s "https://api.github.com/repos/$REPO_OWNER/$REPO_NAME/releases/latest")
        DOWNLOAD_URL=$(echo "$LATEST_RELEASE_INFO" | grep "browser_download_url.*linux-amd64.tar.gz" | cut -d '"' -f 4)
        ACTUAL_VERSION=$(echo "$LATEST_RELEASE_INFO" | grep '"tag_name"' | cut -d '"' -f 4)
        print_status "Latest version found: $ACTUAL_VERSION"
    else
        DOWNLOAD_URL="https://github.com/$REPO_OWNER/$REPO_NAME/releases/download/$VERSION/arcanas-${VERSION#v}-linux-amd64.tar.gz"
        ACTUAL_VERSION="$VERSION"
    fi
    
    # If we got the actual version from API, use it to construct the correct URL
    if [ "$VERSION" = "latest" ] && [ -n "$ACTUAL_VERSION" ]; then
        DOWNLOAD_URL="https://github.com/$REPO_OWNER/$REPO_NAME/releases/download/$ACTUAL_VERSION/arcanas-${ACTUAL_VERSION#v}-linux-amd64.tar.gz"
    fi
    
    if [ -z "$DOWNLOAD_URL" ]; then
        print_error "Could not find download URL for version $VERSION"
        print_error "Make sure the repository exists and has releases"
        exit 1
    fi
    
    print_status "Downloading from: $DOWNLOAD_URL"
    
    # Validate URL format
    if [[ ! "$DOWNLOAD_URL" =~ ^https://github\.com/.*/releases/download/.*\.tar\.gz$ ]]; then
        print_error "Invalid download URL format: $DOWNLOAD_URL"
        exit 1
    fi
    
    wget -q --show-progress "$DOWNLOAD_URL" -O arcanas.tar.gz
    
    if [ ! -f arcanas.tar.gz ]; then
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
    
    # Check which command is available (check full paths for Debian)
    if command -v /usr/sbin/adduser &> /dev/null || command -v adduser &> /dev/null; then
        # Debian/Ubuntu style
        if command -v /usr/sbin/adduser &> /dev/null; then
            /usr/sbin/adduser --system --no-create-home --shell /bin/false --group $SERVICE_USER
        else
            adduser --system --no-create-home --shell /bin/false --group $SERVICE_USER
        fi
    elif command -v /usr/sbin/useradd &> /dev/null || command -v useradd &> /dev/null; then
        # RHEL/CentOS/Fedora style
        if command -v /usr/sbin/useradd &> /dev/null; then
            /usr/sbin/useradd -r -s /bin/false $SERVICE_USER
        else
            useradd -r -s /bin/false $SERVICE_USER
        fi
    else
        print_error "Neither adduser nor useradd command found"
        exit 1
    fi
    
    print_success "Service user created"
}

# Function to install files
install_files() {
    print_status "Installing Arcanas files..."
    
    # Check if this is an update
    if [ -f "$INSTALL_DIR/arcanas" ]; then
        print_status "Arcanas already installed, performing update..."
        
        # Stop the service
        if systemctl is-active --quiet $SERVICE_NAME; then
            print_status "Stopping Arcanas service..."
            systemctl stop $SERVICE_NAME
        fi
        
        # Backup existing binary
        cp "$INSTALL_DIR/arcanas" "$INSTALL_DIR/arcanas.backup"
        print_status "Backed up existing binary"
        
        # Extract new archive
        cd /tmp
        tar -xzf arcanas.tar.gz
        
        # Replace only the binary
        cp arcanas/arcanas "$INSTALL_DIR/arcanas"
        
        # Update static files if they exist
        if [ -d "arcanas/static" ]; then
            rm -rf "$INSTALL_DIR/static"
            cp -r arcanas/static "$INSTALL_DIR/"
            print_status "Updated frontend files"
        fi
        
        # Set ownership and permissions
        chown $SERVICE_USER:$SERVICE_USER "$INSTALL_DIR/arcanas"
        chmod +x "$INSTALL_DIR/arcanas"
        
        # Always update systemd service to apply any changes
        print_status "Updating systemd service..."
        create_systemd_service
        
        # Cleanup
        rm -rf /tmp/arcanas*
        
        print_success "Arcanas updated successfully"
    else
        # Fresh installation
        print_status "Performing fresh installation..."
        
        # Create installation directory
        mkdir -p $INSTALL_DIR
        
        # Extract archive
        cd /tmp
        tar -xzf arcanas.tar.gz
        
        # Copy files to installation directory
        cp -r arcanas/* $INSTALL_DIR/
        
        # Set ownership and permissions
        chown -R $SERVICE_USER:$SERVICE_USER $INSTALL_DIR
        chmod +x $INSTALL_DIR/arcanas
        
        # Cleanup
        rm -rf /tmp/arcanas*
        
        print_success "Files installed to $INSTALL_DIR"
    fi
}

# Function to create systemd service
create_systemd_service() {
    print_status "Creating systemd service..."
    
    cat > /etc/systemd/system/${SERVICE_NAME}.service << EOF
[Unit]
Description=Arcanas
After=network.target

[Service]
Type=simple
User=$SERVICE_USER
Group=$SERVICE_USER
WorkingDirectory=$INSTALL_DIR
ExecStart=$INSTALL_DIR/arcanas
Restart=always
RestartSec=5
Environment=API_PORT=4000

# Security settings
PrivateTmp=false
ProtectSystem=false
ProtectHome=true
ReadWritePaths=$INSTALL_DIR /home/arcanas/data /run/sudo /tmp

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
    print_status "Enabling and starting Arcanas service..."
    
    systemctl enable $SERVICE_NAME
    systemctl start $SERVICE_NAME
    
    # Wait a moment and check status
    sleep 2
    if systemctl is-active --quiet $SERVICE_NAME; then
        print_success "Arcanas service is running"
        print_status "Access at: http://$(hostname -I | awk '{print $1}'):4000"
    else
        print_error "Failed to start Arcanas service"
        systemctl status $SERVICE_NAME
        exit 1
    fi
}

# Function to show post-install info
show_info() {
    echo ""
    if [ -f "$INSTALL_DIR/arcanas.backup" ]; then
        print_success "Arcanas update completed successfully!"
        echo ""
        echo "Backup created: $INSTALL_DIR/arcanas.backup"
    else
        print_success "Arcanas installation completed successfully!"
    fi
    echo ""
    echo "Service Management:"
    echo "  Start:   sudo systemctl start $SERVICE_NAME"
    echo "  Stop:    sudo systemctl stop $SERVICE_NAME"
    echo "  Restart: sudo systemctl restart $SERVICE_NAME"
    echo "  Status:  sudo systemctl status $SERVICE_NAME"
    echo "  Logs:    sudo journalctl -u $SERVICE_NAME -f"
    echo ""
    echo "Access Arcanas at:"
    echo "  http://$(hostname -I | awk '{print $1}'):4000"
    echo ""
    echo "Storage Features:"
    echo "  - Storage pools created in /data/"
    echo "  - Supports MergerFS, LVM, and bind mounts"
    echo "  - RAID array creation and management"
    echo "  - Sudoers configured for storage operations"
    echo ""
    echo "Installation directory: $INSTALL_DIR"
    echo "Installed version: $ACTUAL_VERSION"
    echo ""
}

# Main installation flow
main() {
    detect_os
    install_dependencies
    setup_storage_sudoers
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
