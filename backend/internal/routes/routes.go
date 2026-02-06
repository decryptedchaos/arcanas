/*
 * This file is part of the Arcanas project.
 * Licensed under the Mozilla Public License, v. 2.0.
 * See the LICENSE file at the project root for details.
 */

package routes

import (
	"net/http"
	"strings"

	"arcanas/internal/handlers"
)

func SetupRoutes() *http.ServeMux {
	mux := http.NewServeMux()

	// Authentication endpoints
	mux.HandleFunc("/api/auth/login", handlers.Login)
	mux.HandleFunc("/api/auth/logout", handlers.Logout)
	mux.HandleFunc("/api/auth/validate", handlers.ValidateToken)

	// Hello endpoint (existing)
	mux.HandleFunc("/api/hello", handlers.Hello)

	// Disk storage endpoints
	mux.HandleFunc("/api/disk-stats", handlers.GetDiskStats)
	mux.HandleFunc("/api/disk/smart", handlers.GetSmartStatus)
	mux.HandleFunc("/api/disk/partitions", handlers.GetPartitions)
	mux.HandleFunc("/api/disk/format", handlers.FormatDisk)

	// SMART management endpoints
	mux.HandleFunc("/api/smart/status", handlers.GetAllSmartStatus)
	mux.HandleFunc("/api/smart/test", handlers.RunSmartTest)
	mux.HandleFunc("/api/smart/attributes", handlers.GetSmartAttributes)
	mux.HandleFunc("/api/smart/errors", handlers.GetSmartErrors)
	mux.HandleFunc("/api/smart/setting", handlers.SetSmartSetting)

	// RAID arrays endpoints
	mux.HandleFunc("/api/raid-arrays", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handlers.GetRAIDArrays(w, r)
		case http.MethodPost:
			handlers.CreateRAIDArray(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/api/raid-arrays/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodDelete:
			handlers.DeleteRAIDArray(w, r)
		case http.MethodPost:
			handlers.AddDiskToRAID(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// RAID superblock wipe endpoint (for orphaned RAID metadata)
	mux.HandleFunc("/api/raid-wipe", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			handlers.WipeRAIDSuperblock(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/api/raid-examine", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handlers.ExamineRAIDDevice(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Device mount management endpoints (for unmounting/remounting raw devices for iSCSI)
	mux.HandleFunc("/api/device-mounts", handlers.GetDeviceMounts)
	mux.HandleFunc("/api/device-mounts/unmount", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			handlers.UnmountDevice(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/api/device-mounts/mount", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			handlers.MountDevice(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Storage pools endpoints
	mux.HandleFunc("/api/storage-pools", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handlers.GetStoragePools(w, r)
		case http.MethodPost:
			handlers.CreateStoragePool(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/api/storage-pools/", func(w http.ResponseWriter, r *http.Request) {
		// Check if this is a cleanup endpoint
		// TODO: Remove this cleanup endpoint after migration period (v1.0.0 or later)
		// DEPRECATED: This is temporary migration helper code
		if strings.HasPrefix(r.URL.Path, "/api/storage-pools/cleanup/") {
			handlers.CleanupLegacyPool(w, r)
			return
		}
		// Check if this is a mount/unmount endpoint
		if strings.HasSuffix(r.URL.Path, "/mount") {
			if r.Method == http.MethodPost {
				handlers.MountStoragePool(w, r)
				return
			}
		}
		if strings.HasSuffix(r.URL.Path, "/unmount") {
			if r.Method == http.MethodPost {
				handlers.UnmountStoragePool(w, r)
				return
			}
		}
		// Check if this is an export-mode endpoint
		if strings.HasSuffix(r.URL.Path, "/export-mode") {
			if r.Method == http.MethodPost || r.Method == http.MethodPut {
				handlers.SetPoolExportMode(w, r)
				return
			}
		}

		switch r.Method {
		case http.MethodPut:
			handlers.UpdateStoragePool(w, r)
		case http.MethodDelete:
			handlers.DeleteStoragePool(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// SCSI targets endpoints
	mux.HandleFunc("/api/scsi-targets", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handlers.GetSCSITargets(w, r)
		case http.MethodPost:
			handlers.CreateSCSITarget(w, r)
		case http.MethodDelete:
			handlers.DeleteSCSITarget(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/api/scsi-targets/", func(w http.ResponseWriter, r *http.Request) {
		// Check if this is a backing-stores endpoint
		if strings.HasSuffix(r.URL.Path, "/backing-stores") {
			handlers.GetAvailableBackingStores(w, r)
			return
		}

		switch r.Method {
		case http.MethodPut:
			handlers.UpdateSCSITarget(w, r)
		case http.MethodDelete:
			handlers.DeleteSCSITarget(w, r)
		case http.MethodPost:
			handlers.ToggleSCSITarget(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/api/scsi-targets/sessions", handlers.GetSCSISessions)

	// LUN management endpoints
	mux.HandleFunc("/api/scsi-targets/luns", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			handlers.CreateLUN(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/api/scsi-targets/luns/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodDelete:
			handlers.DeleteLUN(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// ACL management endpoints
	mux.HandleFunc("/api/scsi-targets/acls", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			handlers.CreateACL(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/api/scsi-targets/acls/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodDelete:
			handlers.DeleteACL(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// NEW iSCSI API (Single Target Model)
	// /api/iscsi/target - Get the single iSCSI target
	mux.HandleFunc("/api/iscsi/target", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handlers.GetISCSITarget(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// /api/iscsi/luns - Get all LUNs or create a new LUN
	mux.HandleFunc("/api/iscsi/luns", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handlers.GetISCSILUNs(w, r)
		case http.MethodPost:
			handlers.CreateISCSILUN(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// /api/iscsi/luns/{lun} - Delete a specific LUN
	mux.HandleFunc("/api/iscsi/luns/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodDelete:
			handlers.DeleteISCSILUN(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// /api/iscsi/backends - Get available backend options
	mux.HandleFunc("/api/iscsi/backends", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handlers.GetLUNBackends(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// /api/iscsi/acls - Get all ACLs or create a new ACL
	mux.HandleFunc("/api/iscsi/acls", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handlers.GetISCSIACLs(w, r)
		case http.MethodPost:
			handlers.CreateISCSIACL(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/api/iscsi/acls/", func(w http.ResponseWriter, r *http.Request) {
		// Check if this is a LUN mapping endpoint
		if strings.HasSuffix(r.URL.Path, "/luns") {
			if r.Method == http.MethodPost {
				handlers.MapLUNToACL(w, r)
				return
			}
		}
		// Check if this is a LUN unmapping endpoint
		if strings.Contains(r.URL.Path, "/luns/") {
			if r.Method == http.MethodDelete {
				handlers.UnmapLUNFromACL(w, r)
				return
			}
		}
		// Default ACL operations (delete)
		switch r.Method {
		case http.MethodDelete:
			handlers.DeleteISCSIACL(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/api/iscsi/acls-for-lun", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handlers.GetACLsForLUN(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Volume Group endpoints (for iSCSI LVM management)
	mux.HandleFunc("/api/volume-groups", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handlers.GetVolumeGroups(w, r)
		case http.MethodPost:
			handlers.CreateVolumeGroup(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/api/volume-groups/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodDelete:
			handlers.DeleteVolumeGroup(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/api/volume-groups/available-devices", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handlers.GetVGDevices(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Logical Volume endpoints (for creating/managing LVs from VGs)
	mux.HandleFunc("/api/logical-volumes", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handlers.GetLogicalVolumes(w, r)
		case http.MethodPost:
			handlers.CreateLogicalVolume(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/api/logical-volumes/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodDelete:
			handlers.DeleteLogicalVolume(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/api/logical-volumes/mount", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			handlers.MountLVAsPool(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Samba shares endpoints
	mux.HandleFunc("/api/samba-shares", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handlers.GetSambaShares(w, r)
		case http.MethodPost:
			handlers.CreateSambaShare(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/api/samba-shares/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPut:
			handlers.UpdateSambaShare(w, r)
		case http.MethodDelete:
			handlers.DeleteSambaShare(w, r)
		case http.MethodPost:
			handlers.ToggleSambaShare(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/api/samba-shares/connections", handlers.GetSambaConnections)
	mux.HandleFunc("/api/samba-shares/test", handlers.TestSambaConfig)

	// User management endpoints
	mux.HandleFunc("/api/users", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handlers.GetUsers(w, r)
		case http.MethodPost:
			handlers.CreateUser(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/api/users/", func(w http.ResponseWriter, r *http.Request) {
		// Extract username from URL path
		path := strings.TrimPrefix(r.URL.Path, "/api/users/")
		parts := strings.Split(path, "/")

		// Check if this is a services endpoint
		if len(parts) >= 2 && parts[1] == "services" {
			if r.Method == http.MethodPut {
				handlers.UpdateUserServices(w, r)
				return
			}
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Regular user operations (PUT/DELETE)
		switch r.Method {
		case http.MethodPut:
			handlers.UpdateUser(w, r)
		case http.MethodDelete:
			handlers.DeleteUser(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// NFS exports endpoints
	mux.HandleFunc("/api/nfs-exports", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handlers.GetNFSExports(w, r)
		case http.MethodPost:
			handlers.CreateNFSExport(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/api/nfs-exports/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPut:
			handlers.UpdateNFSExport(w, r)
		case http.MethodDelete:
			handlers.DeleteNFSExport(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/api/nfs-exports/status", handlers.GetNFSExportStatus)
	mux.HandleFunc("/api/nfs-exports/reload", handlers.ReloadNFSConfig)

	// System statistics endpoints
	mux.HandleFunc("/api/system/overview", handlers.GetSystemStats)
	mux.HandleFunc("/api/system/cpu", handlers.GetCPUStats)
	mux.HandleFunc("/api/system/memory", handlers.GetMemoryStats)
	mux.HandleFunc("/api/system/network", handlers.GetNetworkStats)
	mux.HandleFunc("/api/system/storage-health", handlers.GetStorageHealth)
	mux.HandleFunc("/api/system/processes", handlers.GetSystemProcesses)
	mux.HandleFunc("/api/system/logs", handlers.GetSystemLogs)
	mux.HandleFunc("/api/system/reboot", handlers.RebootSystem)
	mux.HandleFunc("/api/system/shutdown", handlers.ShutdownSystem)

	// I/O rate endpoints
	mux.HandleFunc("/api/system/disk-io", handlers.GetDiskIORates)
	mux.HandleFunc("/api/system/array-io", handlers.GetArrayIORates)
	mux.HandleFunc("/api/system/network-io", handlers.GetNetworkIORates)

	// Settings endpoints
	mux.HandleFunc("/api/settings", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handlers.GetSystemSettings(w, r)
		case http.MethodPut:
			handlers.UpdateSystemSettings(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/api/settings/network", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handlers.GetNetworkConfig(w, r)
		case http.MethodPut:
			handlers.UpdateNetworkConfig(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/api/settings/timezone", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handlers.GetTimezone(w, r)
		case http.MethodPut:
			handlers.UpdateTimezone(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	return mux
}
