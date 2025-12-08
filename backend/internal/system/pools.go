package system

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"arcanas/internal/models"
)

func GetStoragePools() ([]models.StoragePool, error) {
	var pools []models.StoragePool

	// Get mergerfs mounts (JBOD pools)
	cmd := exec.Command("findmnt", "-t", "fuse.mergerfs", "-J")
	output, err := cmd.Output()
	if err == nil {
		var mounts []map[string]interface{}
		json.Unmarshal(output, &mounts)

		for _, mount := range mounts {
			pool := models.StoragePool{
				Name:       mount["target"].(string),
				Type:       "mergerfs",
				MountPoint: mount["target"].(string),
				State:      "active",
				CreatedAt:  time.Now(),
			}

			// Parse devices from mount options
			if sources, ok := mount["sources"].([]interface{}); ok {
				for _, source := range sources {
					pool.Devices = append(pool.Devices, source.(string))
				}
			}

			// Get size and usage
			pool.Size, pool.Used = getPoolMountUsage(pool.MountPoint)

			pools = append(pools, pool)
		}
	}

	// TODO: Add other pool types (LVM, etc.)

	return pools, nil
}

func CreateStoragePool(req models.StoragePoolCreateRequest) error {
	switch req.Type {
	case "jbod", "mergerfs":
		return createMergerFSPool(req)
	default:
		return fmt.Errorf("unsupported pool type: %s", req.Type)
	}
}

func createMergerFSPool(req models.StoragePoolCreateRequest) error {
	// Create mount point
	mountPoint := "/mnt/" + req.Name
	cmd := exec.Command("mkdir", "-p", mountPoint)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create mount point: %v", err)
	}

	// Build mergerfs command
	devicesStr := strings.Join(req.Devices, ":")
	config := req.Config
	if config == "" {
		config = "defaults" // Default mergerfs options
	}

	// Mount with mergerfs
	cmd = exec.Command("mergerfs", devicesStr, mountPoint, "-o", config)
	if err := cmd.Run(); err != nil {
		// Cleanup mount point on failure
		exec.Command("rmdir", mountPoint).Run()
		return fmt.Errorf("failed to mount mergerfs: %v", err)
	}

	// Add to fstab for persistence
	fstabEntry := fmt.Sprintf("%s %s fuse.mergerfs %s 0 0\n", devicesStr, mountPoint, config)
	cmd = exec.Command("sh", "-c", fmt.Sprintf("echo '%s' >> /etc/fstab", fstabEntry))
	cmd.Run()

	return nil
}

func UpdateStoragePool(name string, req models.StoragePoolCreateRequest) error {
	// For mergerfs, we need to remount with new config
	mountPoint := "/mnt/" + name

	// Unmount first
	cmd := exec.Command("umount", mountPoint)
	cmd.Run()

	// Remount with new config
	devicesStr := strings.Join(req.Devices, ":")
	config := req.Config
	if config == "" {
		config = "defaults"
	}

	cmd = exec.Command("mergerfs", devicesStr, mountPoint, "-o", config)
	return cmd.Run()
}

func DeleteStoragePool(name string) error {
	mountPoint := "/mnt/" + name

	// Unmount
	cmd := exec.Command("umount", mountPoint)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to unmount: %v", err)
	}

	// Remove from fstab
	cmd = exec.Command("sed", "-i", fmt.Sprintf("\\|%s|d", mountPoint), "/etc/fstab")
	cmd.Run()

	// Remove mount point
	cmd = exec.Command("rmdir", mountPoint)
	return cmd.Run()
}

func FormatDisk(req models.DiskFormatRequest) error {
	// Validate filesystem type
	validFSTypes := map[string]bool{
		"ext4": true, "xfs": true, "btrfs": true,
	}
	if !validFSTypes[req.FSType] {
		return fmt.Errorf("unsupported filesystem type: %s", req.FSType)
	}

	// Format the disk
	var cmd *exec.Cmd
	if req.Label != "" {
		cmd = exec.Command("mkfs."+req.FSType, "-L", req.Label, req.Device)
	} else {
		cmd = exec.Command("mkfs."+req.FSType, req.Device)
	}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to format disk: %v", err)
	}

	return nil
}

func getPoolMountUsage(mountPoint string) (int64, int64) {
	// Get size and usage using df
	cmd := exec.Command("df", "-B", "1", "--output=size,used", mountPoint)
	output, err := cmd.Output()
	if err != nil {
		return 0, 0
	}

	lines := strings.Split(string(output), "\n")
	if len(lines) < 2 {
		return 0, 0
	}

	fields := strings.Fields(lines[1])
	if len(fields) < 2 {
		return 0, 0
	}

	size, _ := strconv.ParseInt(fields[0], 10, 64)
	used, _ := strconv.ParseInt(fields[1], 10, 64)

	return size, used
}
