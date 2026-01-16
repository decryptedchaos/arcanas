# 🐉 Arcanas
> **The Ultimate Self-Hosted NAS Management System** 🚀

![License: MPL 2.0](https://img.shields.io/badge/License-MPL_2.0-brightgreen.svg)
![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go&logoColor=white)
![SvelteKit](https://img.shields.io/badge/SvelteKit-Latest-FF3E00?logo=svelte&logoColor=white)
![Status](https://img.shields.io/badge/Status-Development-orange)

**Arcanas** is a powerful, modern, and lightning-fast NAS management system designed for the enthusiast. Built with a robust Go backend and a sleek SvelteKit frontend, it delivers a premium experience for managing your storage empire.

#### BE ADVISED, THIS IS UNDER HEAVY DEVELOPMENT, WE WOULD WELCOME ANY FIXES YOU WANT TO CONTRIBUTE 

<img width="1864" height="942" alt="image" src="https://github.com/user-attachments/assets/bb4eb1f2-7e7f-4334-a95f-93786cfe698d" />


## ✨ Features

- **🚀 Single Binary Deployment** - Zero external dependencies purely compiled magic.
- **💎 Embedded Frontend** - A beautiful SvelteKit UI baked right into the binary.
- **⚡ Real-time Monitoring** - Live system vitals, disk I/O, and detailed metrics.
- **🛡️ Storage Pools** - Advanced management for MergerFS, LVM, and bind mounts.
- **📼 RAID Mastery** - Create, manage, and monitor RAID arrays with ease.
- **📂 File Sharing** - Instant NFS & Samba/SMB configuration.
- **🎯 iSCSI Targets** - Professional-grade iSCSI target management.
- **🔄 Hot-Reload Dev** - Blazing fast iteration with Air and Vite.

---

## 🚀 Quick Start

### 🛠️ Development Mode

```bash
# 1. First time setup (unleash the power)
sudo ./arcanas setup

# 2. Ignite the engines
./arcanas start

# 3. Access the dashboard
# Frontend: http://localhost:4000 
# API:      http://localhost:4000/api
```

### 📦 Production Installation

```bash
# Install Arcanas (downloads latest release and sets up systemd service)
curl -fsSL https://raw.githubusercontent.com/decryptedchaos/arcanas/master/install.sh | sudo bash

# Install specific version
curl -fsSL https://raw.githubusercontent.com/decryptedchaos/arcanas/master/install.sh | sudo bash -s -- --version v1.0.0
```

The installer will:
- Download the latest release for your platform
- Install system dependencies (Samba, NFS, MergerFS, LVM, mdadm, etc.)
- Create the `arcanas` service user
- Set up systemd service
- Configure firewall rules
- Start the service automatically

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
├── 🧠 backend/           # High-performance Go core
│   ├── cmd/              # Entry points
│   └── internal/         # Business logic & monitoring
├── 🎨 frontend/          # Beautiful SvelteKit UI
│   ├── src/lib/          # Components & stores
│   └── src/routes/       # Pages & layout
└── 📜 arcanas            # The God Script (Management CLI)
```

## ⚙️ Configuration

- **`API_PORT`** - Default: `4000`
- **Storage** - Pools live in `/srv/`
- **Service** - Managed by systemd (auto-configured by installer)

> **Note:** The installer automatically sets up the systemd service. If you need to manually configure it, the service file is located at `/etc/systemd/system/arcanas.service`.

## 🧪 Testing

We take quality seriously. Check [TESTING.md](TESTING.md) for our comprehensive testing guide.

## 🤝 Contributing

Join the revolution!
1. Fork it 🍴
2. Branch it (`git checkout -b feature/cool-stuff`)
3. Code it 💻
4. Push it 🚀
5. PR it 📥

## 📄 License

This project is licensed under the **Mozilla Public License 2.0**.
See [LICENSE](LICENSE) for details.
