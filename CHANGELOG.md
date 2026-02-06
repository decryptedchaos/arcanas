# Changelog

All notable changes to Arcanas will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.6] - 2026-01-26

### Added
- **Dashboard I/O Gauges** - Real-time Disk I/O and Network I/O visualization on dashboard
  - Circular gauge components showing read/write rates for physical disks
  - Network gauges displaying download/upload speeds in Mbps
  - Replaces SystemStatus component with more actionable performance metrics
- **Storage Builder** - New unified workflow system for storage configuration with flexible entry points
  - **Workflow Selector** (`/storage-builder`) - Central hub with 5 workflow options
    - Complete Storage Setup - Full guided wizard (Disks → RAID → LVM → Pool → Shares)
    - RAID Array - Standalone RAID creation from physical disks
    - LVM Volumes - Redirects to existing `/lvm` page for VG/LV management
    - Storage Pool - Create pools from devices, RAID arrays, or LVM logical volumes
    - Network Shares - Create NFS or Samba shares from existing storage
  - **Individual Workflows** - Each component can be configured independently without forced multi-step process
  - **Storage Pool Types** - Pool creation now supports 3 types:
    - Bind Mount - Single device directly mounted
    - MergerFS - Pool multiple devices together
    - LVM Logical Volume - Mount existing unmounted LVs as storage pools
  - **StorageBuilderCard** - Prominent card on Dashboard and Storage pages for easy workflow access
  - **FirstRunBanner** - Non-intrusive banner for first-time users to launch Storage Builder
- **LVM Volumes Page** - New dedicated page (`/lvm`) for managing LVM volume groups and logical volumes
  - Create/delete volume groups from available physical devices (RAID arrays, disks)
  - Create/delete logical volumes from volume groups
  - View VG size, free space, and usage statistics
  - Expandable VG cards showing associated logical volumes
  - Mount LVs as storage pools directly from the LVM page
- **LVM Frontend API** - New `lvmAPI` with methods for VG and LV management
  - `getVolumeGroups()`, `createVolumeGroup()`, `deleteVolumeGroup()`
  - `getLogicalVolumes()`, `createLogicalVolume()`, `deleteLogicalVolume()`
  - `getAvailableDevices()` for discovering devices suitable for VG creation
- **Frontend Component Architecture** - Refactored storage page into reusable components
  - `DisksTab.svelte` - Disk listing with SMART health indicators
  - `RAIDTab.svelte` - RAID array management with delete functionality
  - `PoolsTab.svelte` - Storage pool management with action callbacks
  - Cleaner separation of concerns, easier testing and maintenance
- **Svelte Expression Conventions** - Added `.claude/svelte-conventions.md` documenting:
  - Variable reference patterns (full object paths, store prefixes)
  - Event handler patterns (arrow functions for parameters)
  - Immutable array/object operations
  - Common pitfalls and how to avoid them
  - Component best practices (props, callbacks, accessibility)

### Changed
- **Dashboard Layout** - Removed SystemStatus component in favor of inline I/O gauges
  - SystemStatus and DashboardStats were redundant (both showing CPU/memory)
  - I/O gauges provide unique value not shown elsewhere on dashboard
- **Storage Page Buttons** - "Create RAID Array" and "Create Storage Pool" now link to dedicated workflow pages instead of opening modals
- **NFS/Samba Share Creation** - Simplified to only show Storage Pools in path selection
  - Removed LVM Logical Volumes dropdown section to eliminate confusion
  - Storage Pools are now the single source of truth for shareable storage
  - LVM volumes must be mounted as pools first before they can be shared
- **Pool Type Validation** - Pool creation uses `lvm` type (instead of `lvm_lv`) for LVM logical volume pools to match backend expectations
- **Navigation** - Added "LVM Volumes" link to sidebar between "Storage" and "Sharing"
- **Removed Deprecated Usage Mode** - Storage pools no longer have export mode selector (file/iSCSI/available options removed from pools UI)
- **Documentation** - Updated CLAUDE.md and README.md with Storage Builder, LVM routes, and component architecture

### Fixed
- **Pool Creation API Error** - Fixed 400 Bad Request when creating LVM-backed pools
  - Changed pool type from `lvm_lv` to `lvm` to match backend validation
  - Backend handler accepts `lvm` as valid pool type for LVM logical volumes
- **LVM Signature Wipe Prompt** - Fixed LV creation failing on existing signatures with "Wipe it? [y/n]" prompt
  - Added both `-W y` (wipesignatures) and `-y` (auto-confirm) flags to `lvcreate` command
  - The `-y` flag is required to auto-answer the signature wipe prompt
  - Allows reusing LV names from previously deleted volumes without manual intervention
- **LVM Free Space Detection** - Fixed "not enough free space in VG (have 0 GB)" error when creating LVs
  - `getVGInfo()` was not trimming the "B" suffix from `vgs --units b` output (e.g., "123456789B")
  - `ParseInt()` was failing to parse strings with suffix, returning 0
  - Now properly strips "B" suffix before parsing, same as `GetVolumeGroups()` function
- **LVM Pool Name Expression Error** - Fixed "pool is not defined" error in LVM page
  - Changed `{pool-name || newLV.name}` to `{newLV.poolName || newLV.name}`
  - The hyphenated name was being interpreted as subtraction (pool minus name) instead of a variable reference
- **LVM Input Pattern Regex Error** - Fixed "character class escape cannot be used in class range" validation error
  - Changed `pattern="[a-z0-9-]+"` to `pattern="[a-z0-9\-]+"` to properly escape hyphen in regex character class
  - Allows hyphens in VG and LV names (e.g., "vg-data", "lv-bank")
- **Storage Page Tab Structure** - Fixed improper if/else block nesting that caused "attempted to close an element that was not open" error
  - Added missing `{/if}` to close the outer `activeTab` if/else chain
  - Properly closes tab content div structure

---

## [1.0.5] - 2026-01-25

### Added
- **LVM Volume Group Management** - Separate VG system from storage pools for iSCSI LUNs
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
- **New iSCSI Architecture** - Redesigned iSCSI with single target, multiple LUNs model (enterprise standard)
- **Storage Pool Export Modes** - Unified storage management with three export modes
  - **File Mode** - Mounted filesystem for NFS/Samba sharing (default)
  - **iSCSI Mode** - Unmounted, available for iSCSI/LVM use
  - **Available Mode** - Unmounted, reserved for future use
- **Export Mode API** - `POST /api/storage-pools/{name}/export-mode` to switch modes

### Changed
- **iSCSI Access Control Model** - Changed from automatic ACL generation to manual per-client ACL management
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
