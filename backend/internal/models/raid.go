/*
 * This file is part of the Arcanas project.
 * Licensed under the Mozilla Public License, v. 2.0.
 * See the LICENSE file at the project root for details.
 */

package models

import "time"

type RAIDArray struct {
	Name         string    `json:"name"`
	Device       string    `json:"device"`       // actual device path like /dev/md0
	Level        string    `json:"level"`       // raid0, raid1, raid5, raid6, raid10
	Devices      []string  `json:"devices"`     // list of device paths
	Size         int64     `json:"size"`        // total size in bytes
	Used         int64     `json:"used"`        // used space in bytes
	State        string    `json:"state"`       // active, degraded, failed, syncing
	Health       int       `json:"health"`      // 0-100 percentage
	MountPoint   string    `json:"mount_point"` // where it's mounted
	UUID         string    `json:"uuid"`        // filesystem UUID
	CreatedAt    time.Time `json:"created_at"`
	LastSync     time.Time `json:"last_sync"`
	SyncProgress float64   `json:"sync_progress"` // 0-100 for rebuilding arrays
}

type RAIDCreateRequest struct {
	Name    string   `json:"name,omitempty"` // Optional: auto-generated if not provided
	Level   string   `json:"level"`
	Devices []string `json:"devices"`
}

type StoragePool struct {
	Name        string    `json:"name"`
	Type        string    `json:"type"`        // jbod, mergerfs, lvm, direct, raid
	Devices     []string  `json:"devices"`     // list of device paths
	Size        int64     `json:"size"`        // total size in bytes
	Used        int64     `json:"used"`        // used space in bytes
	Available   int64     `json:"available"`   // available space in bytes
	State       string    `json:"state"`       // active, inactive, error
	MountPoint  string    `json:"mount_point"` // where it's mounted
	Config      string    `json:"config"`      // mergerfs config options
	ExportMode  string    `json:"export_mode"` // "file", "iscsi", "available"
	CreatedAt   time.Time `json:"created_at"`
}

type StoragePoolCreateRequest struct {
	Name       string   `json:"name"`
	Type       string   `json:"type"`
	Devices    []string `json:"devices"`
	Config     string   `json:"config"`
	ExportMode string   `json:"export_mode,omitempty"` // optional, defaults based on type
}

type DiskFormatRequest struct {
	Device string `json:"device"`
	FSType string `json:"fs_type"` // ext4, xfs, btrfs
	Label  string `json:"label"`
}
