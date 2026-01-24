# Changelog

All notable changes to Arcanas will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- **SSH Access Instruction** - Added `ssh root@192.168.1.140` command to CLAUDE.md for quick access to test server
- **ACL UI Improvements** - Enhanced ACL interface with detailed LUN display and unmap functionality:
  - **LUN Details Display** - Each mapped LUN now shows LUN number, name, and size instead of just vague numbers
  - **Unmap Buttons** - Each mapped LUN has an X button (visible on hover) to unmap it from the client
  - **Better Organization** - Each ACL is in its own card with clear header showing client name and mapped LUN count
  - **Empty State** - Shows "No LUNs mapped to this client" when appropriate
  - **Missing LUN Detection** - Shows "(LUN not found)" for mappings to deleted LUNs

### Fixed
- **iSCSI Slow Link Transfer Behavior** - Configured iSCSI target with conservative burst lengths (`FirstBurstLength=32KB`, `MaxBurstLength=64KB`) to prevent false transfer completion reporting on slow links (50Mbps/5MB/s). The default 256KB MaxBurstLength caused transfers to appear complete instantly while data continued transferring at line speed in the background.
  - **Backend**: `EnsureISCSITargetConfigured()` now sets burst lengths via configfs when creating iSCSI targets
  - **Fresh Installs**: `install.sh` creates `arcanas-iscsi-config.service` to apply settings on boot
  - **Upgrades**: Run `/usr/local/bin/arcanas-iscsi-config.sh` manually or restart the service to apply to existing installations
- **iSCSI LUN Cache Mode** - Configured block backstores to disable thin provisioning unmap emulation (`emulate_tpu=0`) for more predictable I/O behavior.
- **Deploy Script Frontend Sync** - Fixed rsync include rules order to properly sync frontend build files to remote server
- **ACL URL Encoding** - Fixed ACL deletion, LUN mapping, and LUN unmapping to properly URL-decode IQNs from request paths (colons in IQNs like `iqn.2016-04.com.open-iscsi:c8e34e60ec9` were being sent as `%3A`)
- **ACL Creation Auto-Remove LUNs** - New ACLs start with no LUNs mapped (auto-mapped LUNs are automatically removed) so users have explicit control over which LUNs to map
- **ACL IQN Parsing** - Fixed `GetISCSIACLs` to correctly extract IQN from targetcli output (was using `parts[0]` which was `o-`, now uses `parts[1]` for the actual IQN)
- **LUN Detection via ConfigFS** - Fixed `getMappedLUNsForACL` to use configfs link file detection instead of parsing targetcli output for more reliable LUN mapping status
- **LUN Backstore Detection** - Fixed `MapLUNToACL` to use `ls` command instead of `info` command for more reliable backstore path parsing from targetcli output
- **LUN Mapping with Existing Symlinks** - Fixed LUN mapping to detect and reuse existing symlinks instead of trying to create duplicate symlinks
  - `DELETE /api/iscsi/acls/{iqn}/luns/{lun}` - Unmap LUN from ACL
  - `GET /api/iscsi/acls-for-lun?lun=N` - Get all ACLs for a specific LUN
- **Target Configuration Change** - Disabled auto-ACL generation (`generate_node_acls=0`) for manual access control
- **ACL Models** - New `ISCSIACL`, `ISCSIACLMappedLUN`, `ACLCreateRequest`, `ACLMapLUNRequest` models
- **ACL System Functions** - `GetISCSIACLs()`, `CreateISCSIACL()`, `DeleteISCSIACL()`, `MapLUNToACL()`, `UnmapLUNFromACL()`, `GetACLsForLUN()`
- **LUN Number Display Fix** - Fixed regex-based LUN parsing to correctly display LUN numbers (0, 1, 2, etc.) instead of all showing as LUN 0
- **Storage Pool Export Modes** - Unified storage management with three export modes:
  - **File Mode** - Mounted filesystem for NFS/Samba sharing (default)
  - **iSCSI Mode** - Unmounted, available for iSCSI/LVM use
  - **Available Mode** - Unmounted, reserved for future use
- **Export Mode API** - `POST /api/storage-pools/{name}/export-mode` to switch modes
- **New iSCSI Architecture** - Redesigned iSCSI with single target, multiple LUNs model (enterprise standard)
- **LVM Volume Group Management** - Separate VG system from storage pools for iSCSI LUNs:
  - **VG Creation** - Create volume groups from physical devices (RAID arrays, disks)
  - **VG Management Tab** - New tab in Storage page for managing volume groups
  - **Quick VG Creation in iSCSI** - Create volume groups directly when creating LVM LUNs
  - **VG Deletion** - Delete VGs with automatic LUN cleanup
- **LVM LUN Backends** - Create flexible, resizable LUNs from LVM volume groups (recommended for sharing storage)
- **Clear Backend Options** - Three LUN backends with clear descriptions:
  - LVM Volume (Flexible) - Create LUNs of any size from VG. Best for sharing storage among clients
  - Block Device (Dedicated) - Use entire disk/RAID for one LUN. Simple but uses whole device
  - File-Based (Testing) - File on disk as LUN. For testing only. Slower performance
- **New iSCSI UI** - `/scsi` route with simplified LUN management interface
- **Storage Pool Editing** - Edit pool name and mount options without deletion
- **NFS Path Editing** - Change export paths on existing NFS shares
- **Storage Pool Dropdown** - Quick-select available storage pools when creating/editing NFS exports
- **RAID Auto-Naming** - System automatically assigns next available device number (e.g., md0, md1) when creating arrays
- **Visual Warnings** - Confirmation dialogs and warnings when making destructive changes
- **Legacy Mount Detection** - Automatic detection and support for `/mnt/arcanas-disk-*` mounts
- **iSCSI Auto-Unmount** - Mounted devices (including RAID arrays) are automatically unmounted when used as iSCSI backing stores
- **Volume Group API** - `GET/POST /api/volume-groups`, `DELETE /api/volume-groups/{name}`, `GET /api/volume-groups/available-devices`
- **LV Info on LUN Cards** - LVM-backed LUNs now display Logical Volume path and size (queried via `lvs`)

### Changed
- **iSCSI Access Control Model** - Changed from automatic ACL generation to manual per-client ACL management:
  - **Breaking Change** - Existing clients will lose access until ACLs are created and LUNs are explicitly mapped
  - **Benefit** - Each client only sees LUNs explicitly mapped to their ACL (improved security and isolation)
  - **Migration Required** - After upgrade, create ACLs for each client and map their required LUNs
- **Target Attribute** - Set `generate_node_acls=0` to disable automatic ACL generation
- **LUN Number Display** - Fixed regex-based LUN parsing to correctly display LUN numbers (0, 1, 2, etc.) instead of all showing as LUN 0
- **Simplified Storage/iSCSI Workflow** - Single "Storage Pool" concept with export modes replaces confusing multiple storage types
- **Pool Export Mode** - Change pool purpose (File Share ↔ iSCSI) with one API call, automatic mount/unmount
- **iSCSI Model** - From multiple targets (one per client) to single target with multiple LUNs
- **Storage Pool Architecture** - MD RAID devices now mount directly at `/srv/{poolname}` instead of using MergerFS wrapper
- **MergerFS Behavior** - Now only used for aggregating multiple raw physical disks (JBOD), not MD devices
- **Device Selection** - Create pool modal filters to show only available (unmounted) devices
- **Pool Type Detection** - New "legacy" type for old `/mnt/arcanas-disk-*` mounts, "direct" for new `/srv/*` mounts
- **iSCSI Backing Stores Display** - Mounted devices now show as available with "Will be auto-unmounted" message instead of blocked

### Fixed
- **LUN Number Parsing** - Fixed regex-based LUN number extraction in both `getNextAvailableLUN()` and `parseLUNsFromTargetcli()` to correctly display LUN numbers (0, 1, 2, etc.)
- **RAID Array Deletion** - Fixed deletion bug by using device path (`/dev/md0`) instead of mdadm name (`arcanas:1`)
- **RAID Array Unmounting** - Arrays are now unmounted before deletion to prevent "Immutable" errors
- **Volume Group Stats** - Fixed VG used/available space showing 0 by using `--units b` for raw byte values instead of human-readable format
- **RAID Persistent Naming** - Added mdadm.conf management with `--homehost` to prevent md127 rename on reboot
- **Storage Pool Creation** - Fixed device filtering to exclude already-mounted devices
- **MD Device Visibility** - Legacy MD device mounts now appear in storage pools API for NFS/Samba sharing
- **NFS Path Update Logic** - Export path changes now properly handle delete+recreate workflow
- **iSCSI Backing Stores UI** - Enhanced to show device availability status with mount point reasons
- **Samba Share Permissions for /mnt Mounts** - Fixed permission issues for Samba shares on mounts under `/mnt` by setting proper ownership and permissions on intermediate mount points
- **MD Device Samba Permissions** - MD RAID devices mounted anywhere are now automatically detected and given proper permissions for Samba access (nobody:nogroup ownership, 0777 permissions)
- **iSCSI Target Creation** - Fixed bug where created targets had no backing store or LUN attached
- **iSCSI Target Update** - Targets can now be updated with new backing stores (replaces existing LUN)
- **iSCSI targetcli Commands** - Fixed all targetcli command paths from relative (`iscsi/`) to absolute (`/iscsi/`) syntax
- **iSCSI LUN Parsing** - Improved LUN detection and backstore path parsing from targetcli output
- **iSCSI ACL LUN Mapping** - Fixed LUN to ACL mapping by dynamically querying backstore path instead of hardcoding `/backstores/block/`
- **ACL IQN Validation** - Fixed IQN input regex pattern to properly validate IQN format with hyphens and colons

---

## [1.0.4] - 2025-01-16

### Added
- **Authentication System** - Login page with JWT-based session management
- **SMART Monitoring** - dedicated page for disk health monitoring with SMART attributes and error logs
- **Settings Page** - System and network settings configuration interface
- **NFS Exports** - Management interface for configuring NFS shares
- **User Management** - User and service account permissions
- **Theme Variables** - CSS custom properties for consistent theming
- **Gauge Component** - Reusable gauge visualization component

### Changed
- **Visual Style** - UI beautification across all pages
- **NFS Exports Backend** - Enhanced handlers and models for NFS export management
- **Disk Handlers** - Expanded disk management capabilities
- **System Page** - Improved system monitoring display
- **Network Settings** - New network configuration options
- **Install Script** - Fixed installation bugs and improved instructions
- **Documentation** - Updated README with better descriptions and installation instructions

### Fixed
- **Download Link** - Fixed download links in README
- **Pipe Display** - Fixed pipe character display issues
- **Install Bug** - Fixed installation script errors

---

## [1.0.3] - 2026-01-15

### Added
- **Theme Update** - Visual theme improvements
- **Automatic Version Bumps** - CI/CD integration for version management

---

## [1.0.2] - 2026-01-15

### Added
- **Visual Beautification** - UI enhancements across storage and system pages

---

## [1.0.1] - 2025-12-10

### Changed
- Rapid development iterations with multiple improvements

---

## [1.0.0] - 2025-12-08

### Added
- **Initial Release** - Storage pools, RAID, NFS, Samba, system monitoring

---

## Version History

| Version | Date | Description |
|---------|------|-------------|
| 1.0.5 | TBD | Storage pool editing, NFS path editing, RAID fixes, enhanced backing stores |
| 1.0.4 | 2025-01-16 | Authentication, SMART monitoring, settings page, NFS exports |
| 1.0.3 | 2026-01-15 | Theme update and automatic version bumps |
| 1.0.2 | 2026-01-15 | Visual beautification |
| 1.0.1 | 2025-12-10 | Rapid development iterations |
| 1.0.0 | 2025-12-08 | Initial release |

---

## Breaking Changes

### v1.0.4 → v1.0.5 (Unreleased)
- **iSCSI Access Control**: Manual ACL management is now required for client access
  - Existing iSCSI clients will lose access to all LUNs after upgrade
  - ACLs must be created for each client IQN that needs access
  - LUNs must be explicitly mapped to each ACL
  - This provides per-client LUN isolation (each client sees only mapped LUNs)
- **Storage Pool Mount Points**: New pools use `/srv/{poolname}` instead of `/mnt/arcanas-disk-*`
- **Legacy pools** are automatically detected and marked as type "legacy"
- To migrate: Delete old pool and recreate with new architecture (data preserved on device)

---

## Upgrade Notes

### From v1.0.3 to v1.0.4
1. Stop the service: `sudo systemctl stop arcanas`
2. Backup your data
3. Download and install the new version
4. Start the service: `sudo systemctl start arcanas`

### From v1.0.4 to v1.0.5 (Unreleased)
1. **iSCSI Migration Required**: After upgrade, existing iSCSI clients will lose access
   - Go to the iSCSI page and create ACLs for each client that needs access
   - For each ACL, map the LUNs that client should be able to access
   - Clients may need to re-discover targets (`iscsiadm -m discovery -t st -p <target>`)
2. Storage pools created with old architecture will continue to work
3. New pools will use direct mount architecture
4. To migrate an old pool: Delete and recreate (data stays on device)

---

## Contributors

- @decryptedchaos - Creator and maintainer

---

## License

This project is licensed under the Mozilla Public License 2.0.
