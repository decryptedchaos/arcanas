# Arcanas - NAS Storage Management System

A powerful, self-hosted NAS management system with a modern web interface.

## Quick Start

### Development Mode

```bash
# First time setup (requires sudo)
sudo ./arcanas setup

# Start development servers
./arcanas start

# Access the application (development mode)
# Frontend: http://localhost:4000 
# API:      http://localhost:4000/api
```

### Production Build

```bash
# Build single binary (frontend embedded)
./arcanas build

# Or use build.sh directly
./build.sh

# Run the binary
./arcanas

# Access the application (production mode)
# Frontend: http://localhost:4000
# API:      http://localhost:4000/api
```

## Features

- **Single binary deployment** - No external dependencies
- **Embedded frontend** - SvelteKit UI built into the binary
- **Real-time monitoring** - Live system statistics
- **Storage pools** - Support for MergerFS, LVM, and bind mounts
- **RAID management** - Create and manage RAID arrays
- **File sharing** - NFS, Samba/SMB support
- **SCSI targets** - iSCSI target management
- **Hot-reload development** - Fast iteration with Air and Vite

## Management Script

The `./arcanas` script provides a unified interface for all development and build tasks:

### Development Commands

```bash
./arcanas start          # Start backend + frontend dev servers
./arcanas stop           # Stop all dev servers
./arcanas restart        # Restart dev servers
./arcanas status         # Show server status
./arcanas logs           # Show all logs
./arcanas logs backend   # Show backend logs only
./arcanas logs frontend  # Show frontend logs only
```

### Build Commands

```bash
./arcanas build          # Build production binary (default version)
./arcanas build 1.2.0    # Build with specific version
```

### Setup Commands

```bash
sudo ./arcanas setup     # First-time development environment setup
```

## Development Details

### Requirements

- **Go 1.21+** - Backend development
- **Node.js 18+** - Frontend development
- **Air** - Hot-reload for Go (installed automatically)

### Manual Development (without wrapper script)

If you prefer to run services manually:

```bash
# Terminal 1: Backend with hot-reload
cd backend && air

# Terminal 2: Frontend with hot-reload
cd frontend && npm run dev
```

### Project Structure

```
arcanas/
├── backend/           # Go backend
│   ├── cmd/          # Application entry points
│   ├── internal/     # Internal packages
│   │   ├── handlers/ # HTTP handlers
│   │   ├── models/   # Data models
│   │   ├── routes/   # Route definitions
│   │   └── system/   # System operations
│   └── .air.toml     # Air configuration
├── frontend/         # SvelteKit frontend
│   ├── src/
│   │   ├── lib/      # Components and utilities
│   │   └── routes/   # SvelteKit routes
│   └── build/        # Production build output
├── arcanas           # Management script
├── build.sh          # Production build script
└── devSetup.sh       # Development environment setup
```

## Configuration

### Environment Variables

- `API_PORT` - Backend API port (default: 4000)

### Storage

- Storage pools are created in `/var/lib/arcanas/`
- Development setup configures necessary sudo permissions

## Production Deployment

### Build

```bash
# Build for current platform
./arcanas build 1.0.0

# Or specify target platform
GOOS=linux GOARCH=amd64 ./build.sh
```

### Deploy

```bash
# Copy binary to server
scp arcanas user@server:/opt/arcanas/

# Make executable
ssh user@server "chmod +x /opt/arcanas/arcanas"

# Run
ssh user@server "/opt/arcanas/arcanas"
```

### Run as Service (systemd)

Create `/etc/systemd/system/arcanas.service`:

```ini
[Unit]
Description=Arcanas NAS Management System
After=network.target

[Service]
Type=simple
User=arcanas
WorkingDirectory=/opt/arcanas
ExecStart=/opt/arcanas/arcanas
Restart=on-failure
Environment="API_PORT=4000"

[Install]
WantedBy=multi-user.target
```

Enable and start:

```bash
sudo systemctl enable arcanas
sudo systemctl start arcanas
sudo systemctl status arcanas
```

## System Requirements

### Development

- Linux (recommended) or macOS
- Go 1.21+
- Node.js 18+
- npm or yarn

### Production

- Linux (for full functionality)
- No runtime dependencies (single binary)

## Troubleshooting

### Development

- **Port already in use**: Stop existing servers with `./arcanas stop`
- **Permission denied**: Run `sudo ./arcanas setup` first
- **Air not found**: Script will auto-install on first run
- **Frontend build fails**: Run `cd frontend && npm install`

### Production

- **Permission denied**: `chmod +x arcanas`
- **Port in use**: Change `API_PORT` environment variable
- **Storage pool creation fails**: Ensure `/var/lib/arcanas` exists and is writable

## Testing

See [TESTING.md](TESTING.md) for comprehensive testing guide.

## API Documentation

The API is available at `http://localhost:4000/api` with the following endpoints:

- `/api/system/*` - System statistics and monitoring
- `/api/storage-pools` - Storage pool management
- `/api/raid-arrays` - RAID array management
- `/api/disk-stats` - Disk information and SMART data
- `/api/nfs-exports` - NFS export management
- `/api/samba-shares` - Samba share management
- `/api/scsi-targets` - iSCSI target management

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests (see TESTING.md)
5. Submit a pull request

## License

[Your License Here]
