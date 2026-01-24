/*
 * This file is part of the Arcanas project.
 * Licensed under the Mozilla Public License, v. 2.0.
 * See the LICENSE file at the project root for details.
 */

package models

import "time"

// ISCSITarget represents the iSCSI server target
// There is typically ONE target per NAS server with multiple LUNs
type ISCSITarget struct {
	IQN         string    `json:"iqn"`          // Target IQN (e.g., iqn.2024-01.com.nas:storage)
	Name        string    `json:"name"`         // Human-readable name
	Status      string    `json:"status"`       // active, inactive
	Description string    `json:"description"`  // Optional description
	LUNs        []ISCSILUN `json:"luns"`        // LUNs exported by this target
	Sessions    int       `json:"sessions"`     // Active sessions
	Created     time.Time `json:"created"`
}

// ISCSILUN represents a Logical Unit Number - a block device exported to clients
type ISCSILUN struct {
	ID           int       `json:"id"`
	LUN          int       `json:"lun"`           // LUN number (0, 1, 2, ...)
	Name         string    `json:"name"`          // Human-readable name
	SizeGB       float64   `json:"size_gb"`       // Size in GB
	BackendType  string    `json:"backend_type"`  // "lvm", "block", "fileio"
	BackendPath  string    `json:"backend_path"`  // Targetcli backstore path: /backstores/block/bs_lun0
	LVPath       string    `json:"lv_path"`       // Underlying LV path (for LVM): /dev/vg-nas/client1
	Status       string    `json:"status"`        // active, inactive
	AllowedIQNs  []string  `json:"allowed_iqns"`  // Initiator IQNs allowed to access this LUN
	Created      time.Time `json:"created"`

	// LV stats for LVM-backed LUNs
	VolumeGroup   string    `json:"volume_group,omitempty"`   // Parent VG name
	LVSizeBytes   int64     `json:"lv_size_bytes,omitempty"`  // LV size in bytes
	LVDataPercent float64   `json:"lv_data_percent,omitempty"` // LV data percentage (0-100)
	LVUsedBytes   int64     `json:"lv_used_bytes,omitempty"`  // LV used space in bytes
}

// LUNCreateRequest creates a new LUN with specified backend
type LUNCreateRequest struct {
	Name        string  `json:"name"`        // Human-readable name (e.g., "Client 1 Storage")
	SizeGB      float64 `json:"size_gb"`     // Size in GB
	BackendType string  `json:"backend_type"` // "lvm", "block", "fileio"

	// For LVM backend (Flexible - recommended)
	VolumeGroup string  `json:"volume_group,omitempty"` // e.g., "vg-raid0"

	// For Block backend (Raw device)
	DevicePath  string  `json:"device_path,omitempty"`  // e.g., "/dev/md0"

	// For FileIO backend (File-based - testing only)
	FilePath    string  `json:"file_path,omitempty"`    // Auto-generated if empty

	// Access control
	AllowedIQNs []string `json:"allowed_iqns,omitempty"` // Initiator IQNs allowed (empty = all)
}

// LUNBackendInfo describes available backend options for LUN creation
type LUNBackendInfo struct {
	Type        string   `json:"type"`        // "lvm", "block", "fileio"
	Name        string   `json:"name"`        // Display name
	Description string   `json:"description"` // What it's for
	Available   bool     `json:"available"`   // Whether this backend can be used
	Resources   []string `json:"resources"`   // Available resources (VGs, devices, etc.)
}

// ISCSIACL represents an Access Control List entry for iSCSI initiators
// Each ACL maps a specific initiator IQN to specific LUNs
type ISCSIACL struct {
	InitiatorIQN string   `json:"initiator_iqn"` // Client's IQN (e.g., iqn.1993-08.org.debian:01:abc123)
	Name         string   `json:"name"`          // Human-readable name for this client
	MappedLUNs   []int    `json:"mapped_luns"`   // LUN numbers this client can access
	Created      time.Time `json:"created"`
}

// ISCSIACLMappedLUN represents a LUN mapped to an ACL with its target LUN number
type ISCSIACLMappedLUN struct {
	SourceLUN int `json:"source_lun"` // The actual LUN number (0, 1, 2, ...)
	TargetLUN int `json:"target_lun"` // The LUN number as seen by this client (can be different)
}

// ACLCreateRequest creates a new ACL entry
type ACLCreateRequest struct {
	InitiatorIQN string `json:"initiator_iqn"` // Client's IQN
	Name         string `json:"name"`          // Human-readable name
}

// ACLMapLUNRequest maps a LUN to an ACL
type ACLMapLUNRequest struct {
	SourceLUN int `json:"source_lun"` // The actual LUN number to map
	TargetLUN int `json:"target_lun"` // Optional: LUN number for client (defaults to source)
}

// Legacy ACL for backward compatibility
type ACL struct {
	ID           int    `json:"id"`
	TargetID     int    `json:"target_id"`
	InitiatorIQN string `json:"initiator_iqn"`
	MappedLUNs   string `json:"mapped_luns"`
}

// Session represents an active iSCSI connection
type Session struct {
	ID         int       `json:"id"`
	TargetID   int       `json:"target_id"`
	Initiator  string    `json:"initiator"`
	IP         string    `json:"ip"`
	Connected  time.Time `json:"connected"`
	BytesRead  int64     `json:"bytes_read"`
	BytesWrite int64     `json:"bytes_write"`
}

// Legacy SCSITarget for backward compatibility (deprecated - use ISCSITarget)
type SCSITarget struct {
	ID           int       `json:"id"`
	Name         string    `json:"name"`
	Status       string    `json:"status"`
	Sessions     int       `json:"sessions"`
	LUNCount     int       `json:"lun_count"`
	Size         string    `json:"size"`
	BackingStore string    `json:"backing_store"`
	InitiatorIPs []string  `json:"initiator_ips"`
	ACLs         []ACL     `json:"acls"`
	LUNs         []LUN     `json:"luns"`
	Created      time.Time `json:"created"`
	LastAccess   time.Time `json:"last_access"`
}

// Legacy LUN for backward compatibility (deprecated - use ISCSILUN)
type LUN struct {
	ID          int    `json:"id"`
	TargetID    int    `json:"target_id"`
	LUN         int    `json:"lun"`
	Device      string `json:"device"`
	Size        int64  `json:"size"`
	BackingFile string `json:"backing_file"`
}

// BackingStore represents a device that can be used as an iSCSI backing store
type BackingStore struct {
	Path       string `json:"path"`       // Device path (e.g., /dev/md0)
	Type       string `json:"type"`       // Device type (e.g., raid0, raid1, part, lvm)
	MountPoint string `json:"mount_point"` // Where it's mounted, empty if available
	Size       string `json:"size"`       // Human-readable size if available
	Available  bool   `json:"available"`  // Whether it can be used as a backing store
	Reason     string `json:"reason"`     // Why it's not available (if applicable)
}

// VolumeGroup represents an LVM volume group for iSCSI LUNs
// VGs are separate from storage pools - they're specifically for creating flexible iSCSI LUNs
type VolumeGroup struct {
	Name      string    `json:"name"`       // VG name (e.g., "vg-iscsi")
	Size      int64     `json:"size"`       // Total size in bytes
	Free      int64     `json:"free"`       // Free space in bytes
	Devices   []string  `json:"devices"`    // Physical devices (PVs) in the VG
	LUNCount  int       `json:"lun_count"`  // Number of LUNs created from this VG
	CreatedAt time.Time `json:"created_at"`
}

// VolumeGroupCreateRequest creates a new VG from devices
type VolumeGroupCreateRequest struct {
	Name    string   `json:"name"`    // VG name
	Devices []string `json:"devices"` // Physical device paths (e.g., /dev/md0, /dev/sdb)
}

// LogicalVolume represents an LVM Logical Volume
// LVs are created from VGs and can be mounted as storage pools or used as iSCSI LUNs
type LogicalVolume struct {
	Name        string    `json:"name"`         // LV name (e.g., "lv-data")
	Path        string    `json:"path"`         // Full device path (e.g., /dev/vg-raid/lv-data)
	VGName      string    `json:"vg_name"`      // Parent volume group
	Size        int64     `json:"size"`         // Size in bytes
	Used        int64     `json:"used"`         // Used space in bytes (if mounted)
	MountPoint  string    `json:"mount_point"`  // Where it's mounted (empty if not mounted)
	Available   bool      `json:"available"`    // Whether available for use (not in use)
	UsedFor     string    `json:"used_for"`     // What this LV is used for ("pool", "iscsi", "available")
	CreatedAt   time.Time `json:"created_at"`
}

// LVCreateRequest creates a new Logical Volume from a VG
type LVCreateRequest struct {
	Name       string  `json:"name"`       // LV name (e.g., "lv-data")
	VGName     string  `json:"vg_name"`   // Parent volume group
	SizeGB     float64 `json:"size_gb"`    // Size in GB
	MountAsPool bool    `json:"mount_as_pool"` // Whether to immediately mount as a storage pool
	PoolName   string  `json:"pool_name"`   // Optional: pool name if mounting
}
