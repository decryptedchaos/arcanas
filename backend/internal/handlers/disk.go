package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"os/exec"
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
	// Use df to get mountpoint and filesystem info
	cmd := exec.Command("df", "--output=source,target,fstype")
	output, err := cmd.Output()
	if err != nil {
		return nil
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines[1:] { // Skip header
		fields := strings.Fields(line)
		if len(fields) >= 3 {
			source := fields[0]
			target := fields[1]
			fstype := fields[2]

			// Check if this matches our device (or contains device name)
			if strings.Contains(source, strings.TrimPrefix(device, "/dev/")) || source == device {
				return &MountInfo{
					Mountpoint: target,
					Filesystem: fstype,
				}
			}
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
	json.NewEncoder(w).Encode(disks)
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
	json.NewEncoder(w).Encode(smart)
}

func GetPartitions(w http.ResponseWriter, r *http.Request) {
	disk := r.URL.Query().Get("disk")

	// TODO: Implement actual partition reading using /proc/partitions or df
	log.Printf("Partitions requested for disk: %s", disk)

	partitions := []models.Partition{
		{
			Device:     "/dev/sda1",
			Mountpoint: "/boot",
			Size:       1000000000,
			Used:       500000000,
			Available:  500000000,
			Usage:      50.0,
			Filesystem: "ext4",
		},
		{
			Device:     "/dev/sda2",
			Mountpoint: "/",
			Size:       1999000000000,
			Used:       1199500000000,
			Available:  799500000000,
			Usage:      60.0,
			Filesystem: "ext4",
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(partitions)
}
