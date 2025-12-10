# Arcanas Quick Reference

## 🚀 Quick Start

```bash
# First time setup
sudo ./arcanas setup

# Start development
./arcanas start

# Access
# Frontend: http://localhost:4000
# API:      http://localhost:4000/api
```

## 📋 Common Commands

### Development
```bash
./arcanas start          # Start dev servers
./arcanas stop           # Stop dev servers
./arcanas restart        # Restart dev servers
./arcanas status         # Check server status
./arcanas logs           # View all logs
./arcanas logs backend   # View backend logs only
./arcanas logs frontend  # View frontend logs only
```

### Building
```bash
./arcanas build          # Build production binary
./arcanas build 1.2.0    # Build with version number
./build.sh               # Alternative build method
```

### Running Production
```bash
./arcanas                # Run the binary
API_PORT=8080 ./arcanas  # Run on custom port
```

## 📁 Project Structure

```
arcanas/
├── arcanas              # Management script ⭐
├── build.sh             # Production build
├── devSetup.sh          # First-time setup
├── backend/             # Go backend
│   ├── cmd/server/      # Entry point
│   ├── internal/        # Business logic
│   └── .air.toml        # Hot-reload config
└── frontend/            # SvelteKit frontend
    ├── src/lib/         # Components
    └── src/routes/      # Pages
```

## 🔧 Development Workflow

1. **First Time**
   ```bash
   sudo ./arcanas setup
   ```

2. **Daily Development**
   ```bash
   ./arcanas start
   # Make changes...
   # Backend auto-reloads with Air
   # Frontend auto-reloads with Vite
   ```

3. **View Logs**
   ```bash
   ./arcanas logs
   # or
   tail -f .backend.log
   tail -f .frontend.log
   ```

4. **Stop Servers**
   ```bash
   ./arcanas stop
   ```

## 🏗️ Building & Deploying

### Build
```bash
./arcanas build 1.0.0
```

### Deploy
```bash
scp arcanas user@server:/opt/arcanas/
ssh user@server "chmod +x /opt/arcanas/arcanas && /opt/arcanas/arcanas"
```

## 🧪 Testing

```bash
cd backend
go test ./...              # Run all tests
go test -v ./...           # Verbose output
go test -cover ./...       # With coverage
```

See [TESTING.md](TESTING.md) for detailed testing guide.

## 🐛 Troubleshooting

### Servers won't start
```bash
./arcanas stop           # Stop any running servers
./arcanas start          # Try again
```

### Port already in use
```bash
./arcanas stop           # Stop dev servers
# or
lsof -ti:4000 | xargs kill  # Kill process on port 4000
```

### Permission errors
```bash
sudo ./arcanas setup     # Re-run setup
```

### Air not found
```bash
go install github.com/air-verse/air@latest
export PATH=$PATH:$(go env GOPATH)/bin
```

## 📡 API Endpoints

- `GET  /api/system/overview` - System stats
- `GET  /api/storage-pools` - List storage pools
- `POST /api/storage-pools` - Create storage pool
- `GET  /api/raid-arrays` - List RAID arrays
- `POST /api/raid-arrays` - Create RAID array
- `GET  /api/disk-stats` - Disk information
- `GET  /api/nfs-exports` - NFS exports
- `GET  /api/samba-shares` - Samba shares
- `GET  /api/scsi-targets` - iSCSI targets

## 🎯 Tips

- Use `./arcanas` for all operations - it handles everything
- Logs are saved to `.backend.log` and `.frontend.log`
- Everything runs on port 4000 (backend + frontend + API)
- Air provides instant backend reloads on code changes
- Vite provides instant frontend reloads on code changes

## 📚 Documentation

- [README.md](README.md) - Full documentation
- [TESTING.md](TESTING.md) - Testing guide
- [devSetup.sh](devSetup.sh) - Setup script details
