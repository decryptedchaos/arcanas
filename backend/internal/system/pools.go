/*
 * This file is part of the Arcanas project.
 * Licensed under the Mozilla Public License, v. 2.0.
 * See the LICENSE file at the project root for details.
 */

package system

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"arcanas/internal/models"
)

func GetStoragePools() ([]models.StoragePool, error) {
	var pools []models.StoragePool

	// Get mergerfs mounts (JBOD pools)
	// Get mergerfs mounts (JBOD pools)
	cmd := exec.Command("findmnt", "-t", "fuse.mergerfs", "-J")
	output, err := cmd.Output()
	if err == nil {
		// findmnt -J returns { "filesystems": [ ... ] }
		var result map[string]interface{}
		if err := json.Unmarshal(output, &result); err == nil {
			if filesystems, ok := result["filesystems"].([]interface{}); ok {
				for _, fsRaw := range filesystems {
					mount, ok := fsRaw.(map[string]interface{})
					if !ok {
						continue
					}

					target, _ := mount["target"].(string)

					// Only list arcanas pools
					if !strings.HasPrefix(target, "/var/lib/arcanas/") {
						continue
					}

					pool := models.StoragePool{
						Name:       strings.TrimPrefix(target, "/var/lib/arcanas/"),
						Type:       "mergerfs",
						MountPoint: target,
						State:      "active",
						CreatedAt:  time.Now(),
					}

					// Parse devices from 'source' string (src1:src2) or 'sources' array
					if src, ok := mount["source"].(string); ok {
						pool.Devices = strings.Split(src, ":")
					} else if sources, ok := mount["sources"].([]interface{}); ok {
						for _, source := range sources {
							if s, ok := source.(string); ok {
								pool.Devices = append(pool.Devices, s)
							}
						}
					}

					// Get size and usage
					pool.Size, pool.Used = GetPathUsage(pool.MountPoint)

					pools = append(pools, pool)
				}
			}
		}
	}

	// Check for existing pool directories (even if not mounted)
	arcanasDir := "/var/lib/arcanas"
	if dirEntries, err := os.ReadDir(arcanasDir); err == nil {
		for _, entry := range dirEntries {
			if !entry.IsDir() {
				continue
			}

			poolName := entry.Name()
			poolPath := filepath.Join(arcanasDir, poolName)

			// Skip if already detected as mounted pool
			alreadyMounted := false
			for _, pool := range pools {
				if pool.Name == poolName {
					alreadyMounted = true
					break
				}
			}
			if alreadyMounted {
				continue
			}

			// Add as unmounted pool
			pool := models.StoragePool{
				Name:       poolName,
				Type:       "directory",
				MountPoint: poolPath,
				State:      "inactive",
				CreatedAt:  time.Now(),
			}

			// Try to get size and usage if path exists
			if _, err := os.Stat(poolPath); err == nil {
				pool.Size, pool.Used = GetPathUsage(poolPath)
			}

			pools = append(pools, pool)
		}
	}

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

	// Prepare each disk (Format -> persistent mount)
	var sourcePaths []string
	for _, device := range req.Devices {
		mountPath, err := prepareDiskForPool(device)
		if err != nil {
			return fmt.Errorf("failed to prepare device %s: %v", device, err)
		}
		sourcePaths = append(sourcePaths, mountPath)
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

	// Create mount point for the POOL
	mountPoint := "/var/lib/arcanas/" + req.Name
	cmd = exec.Command("sudo", "mkdir", "-p", mountPoint)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to create pool mount point %s: %v, output: %s", mountPoint, err, string(output))
	}

	// Build mergerfs command using MOUNT POINTS, not raw devices
	devicesStr := strings.Join(sourcePaths, ":")
	config := req.Config
	if config == "" {
		config = "defaults,allow_other,use_ino"
	}

	// Mount with mergerfs using sudo
	cmd = exec.Command("sudo", "mergerfs", devicesStr, mountPoint, "-o", config)
	if output, err := cmd.CombinedOutput(); err != nil {
		// Cleanup mount point on failure
		exec.Command("sudo", "rmdir", mountPoint).Run()
		return fmt.Errorf("failed to mount mergerfs: %v, output: %s", err, string(output))
	}

	// Add pool to fstab for persistence
	fstabEntry := fmt.Sprintf("%s %s fuse.mergerfs %s 0 0\n", devicesStr, mountPoint, config)
	cmd = exec.Command("sudo", "sh", "-c", fmt.Sprintf("echo '%s' >> /etc/fstab", fstabEntry))
	cmd.Run()

	return nil
}

// prepareDiskForPool ensures a device is formatted and mounted, returning its mount point
func prepareDiskForPool(device string) (string, error) {
	// 1. Check if already mounted
	cmd := exec.Command("findmnt", "-n", "-o", "TARGET", "--source", device)
	output, err := cmd.Output()
	if err == nil && len(output) > 0 {
		return strings.TrimSpace(string(output)), nil
	}

	// 2. Format disk (ext4) if needed
	// Note: blindly formatting for now as usually this UI implies "use this disk"
	// Ideally we check if it has a FS, but for a new pool setup on raw disks, format is expected.
	// Use -F to force if it looks like it has a partition table (safe assuming user selected it for use)
	fmt.Printf("Formatting device %s...\n", device)
	formatCmd := exec.Command("sudo", "mkfs.ext4", "-F", device)
	if out, err := formatCmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("mkfs.ext4 failed: %v %s", err, string(out))
	}

	// 3. Create persistent mount point
	// Naming: /mnt/arcanas-disk-{devname} (e.g. sdb)
	devName := strings.TrimPrefix(device, "/dev/")
	mountPath := "/mnt/arcanas-disk-" + devName

	if err := exec.Command("sudo", "mkdir", "-p", mountPath).Run(); err != nil {
		return "", fmt.Errorf("failed to make dir %s: %v", mountPath, err)
	}

	// 4. Mount it
	if out, err := exec.Command("sudo", "mount", device, mountPath).CombinedOutput(); err != nil {
		return "", fmt.Errorf("mount failed: %v %s", err, string(out))
	}

	// 5. Add to fstab (for disk persistence)
	// UUID is safer, but device path used for simplicity consistent with current architecture
	// We'll use UUID if possible in future, for now device path.
	fstabEntry := fmt.Sprintf("%s %s ext4 defaults 0 0\n", device, mountPath)
	exec.Command("sudo", "sh", "-c", fmt.Sprintf("echo '%s' >> /etc/fstab", fstabEntry)).Run()

	return mountPath, nil
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
	poolPath := "/var/lib/arcanas/" + name
	mountPoint := "/mnt/" + name

	// Check if it's currently mounted and unmount if needed
	if isMounted(mountPoint) {
		cmd := exec.Command("umount", mountPoint)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to unmount: %v", err)
		}
	}

	// Remove from fstab if exists
	cmd := exec.Command("sed", "-i", fmt.Sprintf("\\|%s|d", mountPoint), "/etc/fstab")
	cmd.Run()

	// Remove mount point if exists
	cmd = exec.Command("rmdir", mountPoint)
	cmd.Run()

	// Remove the pool directory
	cmd = exec.Command("rm", "-rf", poolPath)
	return cmd.Run()
}

func isMounted(mountPoint string) bool {
	cmd := exec.Command("findmnt", "-n", mountPoint)
	err := cmd.Run()
	return err == nil
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

func GetPathUsage(mountPoint string) (int64, int64) {
	// First try using df for mounted filesystems
	cmd := exec.Command("df", "-B", "1", "--output=size,used", mountPoint)
	output, err := cmd.Output()
	if err == nil {
		lines := strings.Split(string(output), "\n")
		if len(lines) >= 2 {
			fields := strings.Fields(lines[1])
			if len(fields) >= 2 {
				size, _ := strconv.ParseInt(fields[0], 10, 64)
				used, _ := strconv.ParseInt(fields[1], 10, 64)
				return size, used
			}
		}
	}

	// If df fails, try using du for directory sizes
	cmd = exec.Command("du", "-sb", mountPoint)
	output, err = cmd.Output()
	if err != nil {
		return 0, 0
	}

	lines := strings.Split(string(output), "\n")
	if len(lines) == 0 {
		return 0, 0
	}

	fields := strings.Fields(lines[0])
	if len(fields) < 1 {
		return 0, 0
	}

	used, _ := strconv.ParseInt(fields[0], 10, 64)
	// For directories, we don't have a total size, so estimate based on used space
	size := used + (used / 10) // Add 10% overhead estimate

	return size, used
}
