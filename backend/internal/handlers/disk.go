/*
 * This file is part of the Arcanas project.
 * Licensed under the Mozilla Public License, v. 2.0.
 * See the LICENSE file at the project root for details.
 */

package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"arcanas/internal/models"
	"arcanas/internal/system"
)

type MountInfo struct {
	Mountpoint string
	Filesystem string
}

func getMountInfo(device string) *MountInfo {
	// Use lsblk to get mountpoint and filesystem info - more reliable than df
	// lsblk -J -o NAME,MOUNTPOINT,FSTYPE,PATH
	cmd := exec.Command("lsblk", "-J", "-o", "NAME,MOUNTPOINT,FSTYPE,PATH")
	output, err := cmd.Output()
	if err != nil {
		return nil
	}

	type LsblkDevice struct {
		Name       string        `json:"name"`
		Path       string        `json:"path"`
		Mountpoint *string       `json:"mountpoint"`
		Fstype     *string       `json:"fstype"`
		Children   []LsblkDevice `json:"children"`
	}
	type LsblkOutput struct {
		Blockdevices []LsblkDevice `json:"blockdevices"`
	}

	var result LsblkOutput
	if err := json.Unmarshal(output, &result); err != nil {
		return nil
	}

	// Search for matching device and filesystem info
	var searchDevice func(d LsblkDevice) *MountInfo
	searchDevice = func(d LsblkDevice) *MountInfo {
		// Check if this is our device
		if d.Path == device || strings.Contains(d.Path, strings.TrimPrefix(device, "/dev/")) {
			// If this device has mountpoint/fstype, return it
			if d.Fstype != nil || d.Mountpoint != nil {
				fstype := "unknown"
				mountpoint := ""
				if d.Fstype != nil {
					fstype = *d.Fstype
				}
				if d.Mountpoint != nil {
					mountpoint = *d.Mountpoint
				}
				return &MountInfo{
					Mountpoint: mountpoint,
					Filesystem: fstype,
				}
			}
		}

		// Search children (partitions)
		for _, child := range d.Children {
			// For partitions, check if they belong to our device
			if strings.HasPrefix(child.Path, strings.TrimPrefix(device, "/dev/")) ||
				strings.HasPrefix(child.Path, device) {
				fstype := "unknown"
				mountpoint := ""
				if child.Fstype != nil {
					fstype = *child.Fstype
				}
				if child.Mountpoint != nil {
					mountpoint = *child.Mountpoint
				}
				if fstype != "unknown" || mountpoint != "" {
					return &MountInfo{
						Mountpoint: mountpoint,
						Filesystem: fstype,
					}
				}
			}
			if result := searchDevice(child); result != nil {
				return result
			}
		}

		return nil
	}

	for _, blockdev := range result.Blockdevices {
		if result := searchDevice(blockdev); result != nil {
			return result
		}
	}

	return nil
}

// TODO: Rename this function - it returns disk info, not stats
func GetDiskStats(w http.ResponseWriter, r *http.Request) {
	// Get real storage stats
	storageStats, err := system.GetStorageStats()
	if err != nil {
		log.Printf("Error getting storage stats: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Convert storage stats to disk stats format
	var disks []models.DiskStats
	for _, disk := range storageStats.Disks {
		// Calculate usage percentage
		usage := float64(0)
		if disk.Size > 0 {
			usage = float64(disk.Used) / float64(disk.Size) * 100
		}

		// Get filesystem info
		filesystem := "unknown"
		mountpoint := ""

		// Try to get mountpoint and filesystem from df
		if fsInfo := getMountInfo(disk.Device); fsInfo != nil {
			filesystem = fsInfo.Filesystem
			mountpoint = fsInfo.Mountpoint
		}

		disks = append(disks, models.DiskStats{
			Device:     disk.Device,
			Model:      disk.Model,
			Size:       disk.Size,
			Used:       disk.Used,
			Available:  disk.Size - disk.Used,
			Usage:      usage,
			Mountpoint: mountpoint,
			Filesystem: filesystem,
			ReadOnly:   false, // TODO: Implement read-only detection
			Smart: models.SmartInfo{
				Status:      disk.SmartStatus,
				Health:      disk.Health,
				Temperature: int(disk.Temperature),
				PassedTests: 0, // TODO: Implement SMART test count
				FailedTests: 0,
				LastTest:    time.Now(), // TODO: Implement actual test time
			},
		})
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(disks); err != nil {
		http.Error(w, fmt.Sprintf("Failed to encode response: %v", err), http.StatusInternalServerError)
		return
	}
}

func GetSmartStatus(w http.ResponseWriter, r *http.Request) {
	disk := r.URL.Query().Get("disk")

	// TODO: Implement actual SMART data reading using smartctl
	// For now, return mock data but log the requested disk
	log.Printf("SMART status requested for disk: %s", disk)

	smart := models.SmartInfo{
		Status:      "healthy",
		Health:      95,
		Temperature: 42,
		PassedTests: 150,
		FailedTests: 0,
		LastTest:    time.Now().Add(-24 * time.Hour),
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(smart); err != nil {
		http.Error(w, fmt.Sprintf("Failed to encode response: %v", err), http.StatusInternalServerError)
		return
	}
}

func GetPartitions(w http.ResponseWriter, r *http.Request) {
	disk := r.URL.Query().Get("disk")
	if disk == "" {
		// If no disk specified, return all partitions
		log.Printf("No disk specified, returning all partitions")
	}

	// 1. Run lsblk JSON output
	// lsblk -J -b -o NAME,MOUNTPOINT,SIZE,FSTYPE,UUID,PATH <disk>
	args := []string{"-J", "-b", "-o", "NAME,MOUNTPOINT,SIZE,FSTYPE,UUID,PATH"}
	if disk != "" {
		args = append(args, disk)
	}

	cmd := exec.Command("lsblk", args...)
	output, err := cmd.Output()
	if err != nil {
		log.Printf("Error running lsblk: %v", err)
		http.Error(w, "Failed to list partitions", http.StatusInternalServerError)
		return
	}

	// 2. Parse lsblk JSON
	type LsblkDevice struct {
		Name       string        `json:"name"`
		Path       string        `json:"path"`
		Size       interface{}   `json:"size"` // can be string or number? usually number with -b
		Mountpoint *string       `json:"mountpoint"`
		Fstype     *string       `json:"fstype"`
		Children   []LsblkDevice `json:"children"`
	}
	type LsblkOutput struct {
		Blockdevices []LsblkDevice `json:"blockdevices"`
	}

	var result LsblkOutput
	if err := json.Unmarshal(output, &result); err != nil {
		log.Printf("Error parsing lsblk json: %v", err)
		http.Error(w, "Failed to parse partitions", http.StatusInternalServerError)
		return
	}

	// 3. Flatten and Convert to API Model
	var partitions []models.Partition

	var processDevice func(d LsblkDevice)
	processDevice = func(d LsblkDevice) {
		// Check if it's a partition (children usually, or leaf node)
		// Or just return everything that has a path
		if d.Path != "" {
			part := models.Partition{
				Device:     d.Path,
				Filesystem: "",
			}

			if d.Fstype != nil {
				part.Filesystem = *d.Fstype
			}
			if d.Mountpoint != nil {
				part.Mountpoint = *d.Mountpoint

				// Get usage if mounted
				if size, usedResult := system.GetPathUsage(part.Mountpoint); size > 0 {
					part.Size = size
					part.Used = usedResult
					part.Available = size - usedResult
					if size > 0 {
						part.Usage = float64(usedResult) / float64(size) * 100.0
					}
				}
			} else {
				// Parse size from lsblk (bytes)
				// lsblk JSON size is number if -b used, or is it?
				// json decoder will handle float64 for numbers
				// Let's handle it safely
				switch v := d.Size.(type) {
				case float64:
					part.Size = int64(v)
				case string:
					// try parsing
					if s, err := strconv.ParseInt(v, 10, 64); err == nil {
						part.Size = s
					}
				}
			}

			// Only add if it looks like a partition or logical volume (not the disk itself if it has children)
			// But user might want to format the whole disk?
			// GetPartitions implies "parts".
			// Let's include it.
			partitions = append(partitions, part)
		}

		for _, child := range d.Children {
			processDevice(child)
		}
	}

	for _, d := range result.Blockdevices {
		// If disk was specified, lsblk returns just that disk object
		// If we are looking at the ROOT device, we might skip adding it to partitions list unless it IS a partition?
		// For now, process children. If no children, process self.
		if len(d.Children) > 0 {
			for _, child := range d.Children {
				processDevice(child)
			}
		} else {
			// Single device (e.g. partition passed directly or disk with no parts)
			processDevice(d)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(partitions); err != nil {
		http.Error(w, fmt.Sprintf("Failed to encode response: %v", err), http.StatusInternalServerError)
		return
	}
}
