# 🐉 Arcanas
> **The Ultimate Self-Hosted NAS Management System** 🚀

![License: MPL 2.0](https://img.shields.io/badge/License-MPL_2.0-brightgreen.svg)
![Go Version](https://img.shields.io/badge/Go-1.24+-00ADD8?logo=go&logoColor=white)
![SvelteKit](https://img.shields.io/badge/SvelteKit-Latest-FF3E00?logo=svelte&logoColor=white)
![Status](https://img.shields.io/badge/Status-Active-success)

**Arcanas** is a powerful, modern, and lightning-fast NAS management system designed for the enthusiast. Built with a robust Go backend (standard library only) and a sleek SvelteKit 5 frontend, it delivers a premium experience for managing your storage empire.

## ✨ Features

### Storage Management
- **🛠️ Storage Builder** - Unified workflow system with flexible entry points for storage configuration
  - Complete guided wizard (Disks → RAID → LVM → Pools → Shares)
  - Standalone workflows for RAID, LVM, Pools, and Shares
  - Create LVM-backed pools directly from existing logical volumes
- **🛡️ Storage Pools** - Advanced management for MergerFS (JBOD), bind mounts, and LVM-backed pools
- **📦 LVM Volumes** - Dedicated LVM management page for creating volume groups and logical volumes
  - Create VGs from RAID arrays or physical disks
  - Create flexible, resizable LVs from VGs
  - Mount LVs as storage pools for NFS/Samba sharing
  - Full CRUD operations for volume groups and logical volumes
- **📼 RAID Mastery** - Create, manage, and monitor MD RAID arrays (0, 1, 5, 6, 10)
- **💿 Disk Management** - Format disks, manage partitions, view SMART data
- **🔄 Device Mounting** - Mount/unmount devices for switching between storage pools and iSCSI

### File Sharing
- **📂 NFS Exports** - Configure and manage NFS shares with client access rules
- **🤝 Samba/SMB** - Set up Windows-compatible file sharing
- **🎯 Path Editing** - Edit export paths on existing shares without delete/recreate

### iSCSI
- **💾 iSCSI Targets** - Professional-grade iSCSI target management with LIO/targetcli
- **🔌 LUN Management** - Create and manage LUNs with block or fileio backstores
- **🔐 ACL Configuration** - Manage initiator IQNs and access control lists

### Monitoring & System
- **⚡ Real-time Monitoring** - Live CPU, memory, network, and disk I/O metrics with circular gauge visualizations
- **📊 SMART Data** - Drive health monitoring with test execution
- **🖥️ System Info** - View processes, logs, and system resources
- **👥 User Management** - Manage users and service account permissions

### Architecture
- **🚀 Single Binary** - Zero external runtime dependencies, embedded frontend
- **⚡ Hot-Reload Dev** - Blazing fast iteration with Air and Vite 6
- **🎨 Modern UI** - Svelte 5 with SvelteKit, Tailwind CSS, dark mode support

---

## 🚀 Quick Start

### 🛠️ Development Mode

```bash
# 1. Clone the repository
git clone https://github.com/decryptedchaos/arcanas.git
cd arcanas

# 2. Install frontend dependencies
cd frontend && npm install

# 3. Start backend with hot-reload (from project root)
cd ..
./dev.sh

# 4. Or start frontend separately
cd frontend && npm run dev    # Frontend on :5173
cd ../backend && go run cmd/server/main.go  # Backend on :4000
```

**Access the dashboard:**
- Frontend: http://localhost:5173 (dev) or http://localhost:4000 (prod)
- API: http://localhost:4000/api

### 📦 Production Installation

#### Install Latest Release
```bash
curl -fsSL https://raw.githubusercontent.com/decryptedchaos/arcanas/master/install.sh | sudo bash
```

#### Install Specific Version
```bash
curl -fsSL https://raw.githubusercontent.com/decryptedchaos/arcanas/master/install.sh | sudo bash -s -- --version v1.0.0
```

The installer will:
- Download the latest release for your platform (amd64/arm64)
- Install system dependencies (Samba, NFS, MergerFS, LVM, mdadm, targetcli)
- Create the `arcanas` service user and data directories
- Set up systemd service with auto-start
- Configure sudoers for privileged operations
- Start the service

**Access Arcanas:**
```
Frontend: http://your-server-ip:4000
API:      http://your-server-ip:4000/api
```

**Service Management:**
```bash
sudo systemctl start arcanas    # Start the service
sudo systemctl stop arcanas     # Stop the service
sudo systemctl restart arcanas  # Restart the service
sudo systemctl status arcanas   # Check service status
sudo journalctl -u arcanas -f   # View logs
```

---

## 🏗️ Architecture

```
arcanas/
├── 🧠 backend/              # Go backend (standard library only)
│   ├── cmd/server/         # Main entry point
│   ├── internal/
│   │   ├── handlers/       # HTTP request handlers
│   │   ├── system/         # System command execution
│   │   ├── models/         # Data structures
│   │   ├── routes/         # API routing
│   │   └── utils/          # Helper functions
│   └── static/             # Embedded frontend (generated)
├── 🎨 frontend/            # SvelteKit 5 frontend
│   ├── src/lib/           # Components, API client, stores
│   ├── src/routes/        # Pages (storage, scsi, nfs, samba, etc.)
│   └── static/            # Static assets
├── 📜 build.sh             # Production build script
├── 🔄 dev.sh               # Development hot-reload script
└── 🚀 deploy.sh            # Remote deployment script
```

### Storage Architecture

Arcanas provides multiple storage management approaches for different use cases:

**LVM (Logical Volume Manager)** - Flexible volume management
- Create volume groups from RAID arrays or physical disks
- Create logical volumes of any size from VGs (resizable!)
- Mount LVs at `/srv/{poolname}/` for NFS/Samba sharing
- Ideal for: Flexible storage allocation, volume resizing, snapshots

**MergerFS (JBOD)** - Aggregate independent disks
- Multiple physical disks pooled together
- No striping/redundancy (filesystem-level only)
- Mounts at `/srv/{poolname}/`
- Ideal for: Pooling disks of different sizes without RAID

**Bind Mounts** - Share existing directories
- Mount any directory path as a storage pool
- No filesystem creation required
- Ideal for: Reusing existing data locations

**Storage Pool Types:**
| Type | Use Case | Managed Via |
|------|----------|------------|
| LVM LV | Flexible resizable volumes | LVM Volumes page |
| MergerFS | JBOD disk aggregation | Storage Pools page |
| Bind Mount | Share existing paths | Storage Pools page |

---

## ⚙️ Configuration

### Environment Variables
- **`API_PORT`** - Server port (default: `4000`)
- **`DEV_MODE`** - Enable continuous frontend rebuild (default: `false`)

### Storage Locations
- **Storage Pools** - `/srv/{poolname}/`
- **Legacy Mounts** - `/mnt/arcanas-disk-{device}/`
- **iSCSI Storage** - `/var/lib/arcanas/iscsi/`

### System Requirements
- Linux (tested on Arch, Ubuntu, Debian)
- Go 1.24+ (for development)
- Node.js 20+ (for frontend development)
- Sudo access for privileged operations

---

## 🛠️ Development

### Build for Production
```bash
./build.sh    # Builds frontend, embeds in Go binary
# Output: ./arcanas (single binary)
```

### Build Frontend Only
```bash
cd frontend
npm run build    # Outputs to build/
```

### Run Backend (One-Shot)
```bash
cd backend
go run cmd/server/main.go
```

### Run Backend (Auto-Rebuild)
```bash
DEV_MODE=true go run cmd/server/main.go
```

### Frontend Development
```bash
cd frontend
npm run dev      # Start Vite dev server on :5173
npm run build    # Build for production
npm run lint     # Prettier + ESLint check
```

### Remote Deployment
```bash
./deploy.sh root@192.168.1.140    # Deploy to remote server
```

---

## 🔒 Security

- **Sudoers Configuration** - Privileged commands executed via passwordless sudo
- **Path Validation** - All paths validated to prevent traversal attacks
- **Input Sanitization** - User inputs sanitized before system command execution
- **Service Isolation** - Runs as dedicated `arcanas` user

---

## 📄 License

This project is licensed under the **Mozilla Public License 2.0**.
See [LICENSE](LICENSE) for details.

---

## 🤝 Contributing

Contributions are welcome! The project uses MPL 2.0 which allows for:
- Proprietary use of modified files (you keep your modifications private)
- Copyleft on original files (modifications to MPL-licensed files must remain MPL)

Please ensure all Go files include the MPL license header.

---

## 📜 Changelog

See [CHANGELOG.md](CHANGELOG.md) for version history and changes.
