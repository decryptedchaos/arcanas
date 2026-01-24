#!/bin/bash

# Development Setup Script for Arcanas
# Run this once to set up your development environment

set -e

echo "Setting up Arcanas development environment..."

# Check if running as root
if [[ $EUID -ne 0 ]]; then
   echo "This script must be run as root (use sudo)"
   exit 1
fi

# Create data directory
echo "Creating /var/lib/arcanas..."
mkdir -p /var/lib/arcanas
chown -R $SUDO_USER:$SUDO_USER /var/lib/arcanas
chmod 755 /var/lib/arcanas

# Setup sudoers for development
echo "Setting up sudoers configuration..."
mkdir -p /etc/sudoers.d

cat > /etc/sudoers.d/arcanas-storage << 'EOF'
# Arcanas storage operations sudoers configuration
# Allows users in sudo group to run storage commands without password

Cmnd_Alias ARCANAS_STORAGE = /bin/mkdir, /usr/bin/mkdir, /bin/mount, /usr/bin/mount, /bin/umount, /usr/bin/umount, /usr/sbin/vgcreate, /usr/sbin/lvcreate, /sbin/mkfs, /usr/sbin/mkfs*, /usr/bin/mergerfs, /bin/sh, /usr/bin/sh, /usr/bin/sed, /bin/sed, /bin/rmdir, /usr/bin/rmdir, /usr/sbin/vgremove, /usr/sbin/lvremove, /usr/sbin/lvs, /usr/sbin/vgs, /usr/sbin/pvdisplay, /usr/sbin/pvremove, /usr/sbin/pvcreate, /usr/sbin/chown, /usr/bin/chown, /bin/chown, /usr/sbin/mdadm, /usr/bin/mdadm, /usr/bin/true, /usr/sbin/wipefs, /sbin/wipefs, /usr/bin/wipefs, /usr/bin/which, /bin/which, /usr/bin/findmnt, /usr/sbin/findmnt, /sbin/blockdev, /usr/sbin/blockdev

%sudo ALL=(ALL) NOPASSWD: ARCANAS_STORAGE
EOF

chmod 440 /etc/sudoers.d/arcanas-storage

# Validate sudoers file
if ! visudo -c -f /etc/sudoers.d/arcanas-storage; then
    echo "ERROR: Sudoers file is invalid!"
    rm /etc/sudoers.d/arcanas-storage
    exit 1
fi

echo ""
echo "✓ Development environment setup complete!"
echo ""
echo "Data directory: /var/lib/arcanas"
echo "Owner: $SUDO_USER:$SUDO_USER"
echo "Sudoers configured for storage operations"
echo ""
echo "You can now run the application with:"
echo "  cd backend && go run cmd/server/main.go"
echo "  or"
echo "  ./nas-dashboard"
echo ""
