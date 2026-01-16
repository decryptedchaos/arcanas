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

### 📦 Production Build

```bash
# Forge the single binary
./arcanas build

# Deploy the artifact
./arcanas

# Access your NAS
# Frontend: http://localhost:4000
# API:      http://localhost:4000/api
```

---

## 🎮 Management Script

The `./arcanas` script is your command center using it feels like cheating.

| Command | Action |
|---------|--------|
| `./arcanas start` | 🔥 Start backend & frontend dev servers |
| `./arcanas stop` | 🛑 Stop all running servers |
| `./arcanas restart` | 🔄 Restart everything |
| `./arcanas status` | 📊 Check server health |
| `./arcanas logs` | 📜 View live logs |
| `./arcanas build` | 🏗️ Compile production binary |

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
- **Storage** - Pools live in `/var/lib/arcanas/`

## 🐧 Production Deployment (systemd)

Turn Arcanas into a proper system service:

1. Create `/etc/systemd/system/arcanas.service`:
```ini
[Unit]
Description=Arcanas NAS System 🐉
After=network.target

[Service]
Type=simple
User=arcanas
ExecStart=/opt/arcanas/arcanas
Restart=on-failure
Environment="API_PORT=4000"

[Install]
WantedBy=multi-user.target
```

2. Enable & Start:
```bash
sudo systemctl enable --now arcanas
```

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
