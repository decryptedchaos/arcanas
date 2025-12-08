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
				Name:       strings.TrimPrefix(mount["target"].(string), "/var/lib/arcanas/"),
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
	case "bind":
		return createBindMountPool(req)
	case "lvm":
		return createLVMPool(req)
	default:
		return fmt.Errorf("unsupported pool type: %s", req.Type)
	}
}

func createMergerFSPool(req models.StoragePoolCreateRequest) error {
	// Check if mergerfs is available
	if _, err := exec.LookPath("mergerfs"); err != nil {
		return fmt.Errorf("mergerfs is not installed. Install with:\nUbuntu/Debian: sudo apt install mergerfs\nFedora/CentOS: sudo dnf install mergerfs\nArch: sudo pacman -S mergerfs\nOr download from: https://github.com/trapexit/mergerfs/releases")
	}

	// Ensure data directory exists and has proper permissions
	cmd := exec.Command("sudo", "mkdir", "-p", "/var/lib/arcanas")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to create data directory /var/lib/arcanas: %v, output: %s", err, string(output))
	}

	// Set ownership of data directory to arcanas user if it exists
	cmd = exec.Command("sudo", "chown", "-R", "arcanas:arcanas", "/var/lib/arcanas")
	if output, err := cmd.CombinedOutput(); err != nil {
		// Log warning but don't fail if arcanas user doesn't exist
		fmt.Printf("Warning: failed to set data directory ownership: %v, output: %s\n", err, string(output))
	}

	// Create mount point using sudo
	mountPoint := "/var/lib/arcanas/" + req.Name
	cmd = exec.Command("sudo", "mkdir", "-p", mountPoint)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to create mount point %s: %v, output: %s", mountPoint, err, string(output))
	}

	// Build mergerfs command
	devicesStr := strings.Join(req.Devices, ":")
	config := req.Config
	if config == "" {
		config = "defaults" // Default mergerfs options
	}

	// Mount with mergerfs using sudo
	cmd = exec.Command("sudo", "mergerfs", devicesStr, mountPoint, "-o", config)
	if output, err := cmd.CombinedOutput(); err != nil {
		// Cleanup mount point on failure
		exec.Command("sudo", "rmdir", mountPoint).Run()
		return fmt.Errorf("failed to mount mergerfs: %v, output: %s", err, string(output))
	}

	// Add to fstab for persistence using sudo
	fstabEntry := fmt.Sprintf("%s %s fuse.mergerfs %s 0 0\n", devicesStr, mountPoint, config)
	cmd = exec.Command("sudo", "sh", "-c", fmt.Sprintf("echo '%s' >> /etc/fstab", fstabEntry))
	cmd.Run()

	return nil
}

func createBindMountPool(req models.StoragePoolCreateRequest) error {
	if len(req.Devices) != 1 {
		return fmt.Errorf("bind mount pools require exactly one device")
	}

	// Ensure data directory exists and has proper permissions
	cmd := exec.Command("sudo", "mkdir", "-p", "/var/lib/arcanas")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to create data directory /var/lib/arcanas: %v, output: %s", err, string(output))
	}

	// Set ownership of data directory to arcanas user if it exists
	cmd = exec.Command("sudo", "chown", "-R", "arcanas:arcanas", "/var/lib/arcanas")
	if output, err := cmd.CombinedOutput(); err != nil {
		// Log warning but don't fail if arcanas user doesn't exist
		fmt.Printf("Warning: failed to set data directory ownership: %v, output: %s\n", err, string(output))
	}

	// Create mount point using sudo
	mountPoint := "/var/lib/arcanas/" + req.Name
	cmd = exec.Command("sudo", "mkdir", "-p", mountPoint)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create mount point: %v", err)
	}

	// Create bind mount
	cmd = exec.Command("mount", "--bind", req.Devices[0], mountPoint)
	if err := cmd.Run(); err != nil {
		// Cleanup mount point on failure
		exec.Command("rmdir", mountPoint).Run()
		return fmt.Errorf("failed to create bind mount: %v", err)
	}

	// Add to fstab for persistence
	fstabEntry := fmt.Sprintf("%s %s none bind 0 0\n", req.Devices[0], mountPoint)
	cmd = exec.Command("sh", "-c", fmt.Sprintf("echo '%s' >> /etc/fstab", fstabEntry))
	cmd.Run()

	return nil
}

func createLVMPool(req models.StoragePoolCreateRequest) error {
	// Check if LVM tools are available
	if _, err := exec.LookPath("lvcreate"); err != nil {
		return fmt.Errorf("LVM tools not installed. Install with: sudo apt install lvm2")
	}

	if len(req.Devices) == 0 {
		return fmt.Errorf("at least one device is required for LVM")
	}

	// Ensure data directory exists and has proper permissions
	cmd := exec.Command("sudo", "mkdir", "-p", "/var/lib/arcanas")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to create data directory /var/lib/arcanas: %v, output: %s", err, string(output))
	}

	// Create volume group
	vgName := "vg_" + req.Name
	cmd = exec.Command("vgcreate", append([]string{vgName}, req.Devices...)...)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create volume group: %v", err)
	}

	// Create logical volume (use 90% of available space)
	lvName := "lv_" + req.Name
	cmd = exec.Command("lvcreate", "-l", "90%VG", "-n", lvName, vgName)
	if err := cmd.Run(); err != nil {
		// Cleanup volume group on failure
		exec.Command("vgremove", "-f", vgName).Run()
		return fmt.Errorf("failed to create logical volume: %v", err)
	}

	// Create filesystem
	lvPath := fmt.Sprintf("/dev/%s/%s", vgName, lvName)
	cmd = exec.Command("mkfs.ext4", lvPath)
	if err := cmd.Run(); err != nil {
		// Cleanup on failure
		exec.Command("lvremove", "-f", lvPath).Run()
		exec.Command("vgremove", "-f", vgName).Run()
		return fmt.Errorf("failed to create filesystem: %v", err)
	}

	// Create mount point using sudo
	mountPoint := "/var/lib/arcanas/" + req.Name
	cmd = exec.Command("sudo", "mkdir", "-p", mountPoint)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create mount point: %v", err)
	}

	// Mount the logical volume
	cmd = exec.Command("mount", lvPath, mountPoint)
	if err := cmd.Run(); err != nil {
		exec.Command("rmdir", mountPoint).Run()
		return fmt.Errorf("failed to mount logical volume: %v", err)
	}

	// Add to fstab for persistence
	fstabEntry := fmt.Sprintf("%s %s ext4 defaults 0 0\n", lvPath, mountPoint)
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
