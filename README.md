# NAS Dashboard Deployment

## Quick Start (Single Binary)

1. **Build the application:**
   ```bash
   ./build.sh
   ```

2. **Run the binary:**
   ```bash
   ./nas-dashboard
   ```

3. **Access the dashboard:**
   - Web UI: http://localhost:4000
   - API: http://localhost:4000/api

## Features

- **Single binary deployment** - No external dependencies
- **Embedded frontend** - SvelteKit UI built into the binary
- **Cross-platform** - Runs on Linux, macOS, Windows
- **Real-time monitoring** - Live system statistics
- **Zero configuration** - Works out of the box
- **Storage pools** - Created in `/var/lib/arcanas/` automatically

## Development Mode

For development with hot reload:

```bash
# Terminal 1: Backend
cd backend && go run cmd/server/main.go

# Terminal 2: Frontend
cd frontend && npm run dev
```

## Production Deployment

1. **Build for target platform:**
   ```bash
   # Linux
   GOOS=linux GOARCH=amd64 ./build.sh
   
   # macOS
   GOOS=darwin GOARCH=amd64 ./build.sh
   
   # Windows
   GOOS=windows GOARCH=amd64 ./build.sh
   ```

2. **Deploy the binary:**
   ```bash
   scp nas-dashboard user@server:/opt/nas-dashboard/
   ssh user@server "chmod +x /opt/nas-dashboard/nas-dashboard"
   ```

3. **Run as service (optional):**
   ```bash
   sudo cp nas-dashboard.service /etc/systemd/system/
   sudo systemctl enable nas-dashboard
   sudo systemctl start nas-dashboard
   ```

## System Requirements

- Linux, macOS, or Windows
- Go 1.21+ (for building only)
- Node.js 18+ (for building frontend only)
- No runtime dependencies required

## Troubleshooting

- **Permission denied**: `chmod +x nas-dashboard`
- **Port in use**: Change `API_PORT` environment variable
- **Build fails**: Check Node.js and Go versions
- **Storage pool creation fails**: Ensure `/var/lib/arcanas` is writable or run with proper permissions
