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
	// Initialize as empty slice to avoid null JSON encoding
	pools := make([]models.StoragePool, 0)

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
					if !strings.HasPrefix(target, "/srv/") {
						continue
					}

					pool := models.StoragePool{
						Name:       strings.TrimPrefix(target, "/srv/"),
						Type:       "mergerfs",
						MountPoint: target,
						State:      "active",
						CreatedAt:  time.Now(),
					}

					// Parse devices from mergerfs config
					// findmnt may show truncated source (e.g., "b:c"), so read from fstab
					// to get the actual source mount points
					fstabData, err := os.ReadFile("/etc/fstab")
					if err == nil {
						fstabLines := strings.Split(string(fstabData), "\n")
						for _, line := range fstabLines {
							if strings.Contains(line, target) && strings.Contains(line, "fuse.mergerfs") {
								// Parse: /mnt/arcanas-disk-sdb:/mnt/arcanas-disk-sdc /srv/vt fuse.mergerfs ...
								fields := strings.Fields(line)
								if len(fields) >= 1 {
									// First field is the source string
									source := fields[0]
									if source != "" && source != "none" {
										// Split by colon to get individual mount points
										pool.Devices = strings.Split(source, ":")
										break
									}
								}
							}
						}
					}

					// Fallback: if fstab parsing failed, use findmnt output (may be truncated)
					if len(pool.Devices) == 0 {
						if src, ok := mount["source"].(string); ok && src != "" {
							pool.Devices = strings.Split(src, ":")
						} else if sources, ok := mount["sources"].([]interface{}); ok {
							for _, source := range sources {
								if s, ok := source.(string); ok {
									pool.Devices = append(pool.Devices, s)
								}
							}
						}
					}

					// Get size and usage
					pool.Size, pool.Used = GetPathUsage(pool.MountPoint)
					pool.Available = pool.Size - pool.Used

					pools = append(pools, pool)
				}
			}
		}
	}

	// Check for existing pool directories (even if not mounted)
	arcanasDir := "/srv"
	if dirEntries, err := os.ReadDir(arcanasDir); err == nil {
		for _, entry := range dirEntries {
			if !entry.IsDir() {
				continue
			}

			poolName := entry.Name()

			// Skip system directories
			if poolName == "ftp" || poolName == "http" {
				continue
			}

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
				pool.Available = pool.Size - pool.Used
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
	cmd := exec.Command("sudo", "mkdir", "-p", "/srv")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to create data directory /srv: %v, output: %s", err, string(output))
	}

	// Set ownership of data directory to arcanas user if it exists
	cmd = exec.Command("sudo", "chown", "-R", "arcanas:arcanas", "/srv")
	if output, err := cmd.CombinedOutput(); err != nil {
		// Log warning but don't fail if arcanas user doesn't exist
		fmt.Printf("Warning: failed to set data directory ownership: %v, output: %s\n", err, string(output))
	}

	// Create mount point for the POOL
	mountPoint := "/srv/" + req.Name
	cmd = exec.Command("sudo", "mkdir", "-p", mountPoint)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to create pool mount point %s: %v, output: %s", mountPoint, err, string(output))
	}

	// Build mergerfs command using MOUNT POINTS, not raw devices
	// MergerFS requires paths to be separated by colons
	// Use a special syntax to ensure proper parsing: /path1:/path2
	devicesStr := strings.Join(sourcePaths, ":")

	config := req.Config
	if config == "" {
		config = "defaults,allow_other,use_ino"
	}

	// Mount with mergerfs using sudo
	// The source will be displayed as "b:c" in df due to path truncation, but the mount works correctly
	cmd = exec.Command("sudo", "mergerfs", devicesStr, mountPoint, "-o", config)
	if output, err := cmd.CombinedOutput(); err != nil {
		// Cleanup mount point on failure
		exec.Command("sudo", "rmdir", mountPoint).Run()
		return fmt.Errorf("failed to mount mergerfs: %v, output: %s", err, string(output))
	}

	// Verify the mount was successful
	if !isMounted(mountPoint) {
		exec.Command("sudo", "rmdir", mountPoint).Run()
		return fmt.Errorf("mergerfs mount verification failed for %s", mountPoint)
	}

	// Set permissions for the mounted pool so Samba and other services can access it
	// Set ownership to allow Samba to write (nogroup is the default guest account)
	cmd = exec.Command("sudo", "chown", "-R", "nobody:nogroup", mountPoint)
	if output, err := cmd.CombinedOutput(); err != nil {
		// Log warning but don't fail - this might not be critical
		fmt.Printf("Warning: failed to set pool ownership: %v, output: %s\n", err, string(output))
	}

	// Set permissions to allow read/write access
	cmd = exec.Command("sudo", "chmod", "-R", "0777", mountPoint)
	if output, err := cmd.CombinedOutput(); err != nil {
		// Log warning but don't fail
		fmt.Printf("Warning: failed to set pool permissions: %v, output: %s\n", err, string(output))
	}

	// Add pool to fstab for persistence
	fstabEntry := fmt.Sprintf("%s %s fuse.mergerfs %s 0 0\n", devicesStr, mountPoint, config)
	cmd = exec.Command("sudo", "sh", "-c", fmt.Sprintf("echo '%s' >> /etc/fstab", fstabEntry))
	if err := cmd.Run(); err != nil {
		// Log warning but don't fail - fstab update is secondary
		fmt.Printf("Warning: failed to add to fstab: %v\n", err)
	}

	return nil
}

// prepareDiskForPool ensures a device is formatted and mounted, returning its mount point
func prepareDiskForPool(device string) (string, error) {
	//1. Check if already mounted
	cmd := exec.Command("findmnt", "-n", "-o", "TARGET", "--source", device)
	output, err := cmd.Output()
	if err == nil && len(output) > 0 {
		return strings.TrimSpace(string(output)), nil
	}

	//2. Check if device has a filesystem already
	cmd = exec.Command("blkid", "-o", "value", "-s", "TYPE", device)
	fsOutput, err := cmd.Output()
	hasFilesystem := err == nil && len(strings.TrimSpace(string(fsOutput))) > 0

	if hasFilesystem {
		fmt.Printf("Device %s already has filesystem, skipping format\n", device)
	} else {
		// Format disk (ext4) if needed
		// Use -F to force if it looks like it has a partition table
		fmt.Printf("Formatting device %s...\n", device)
		formatCmd := exec.Command("sudo", "mkfs.ext4", "-F", device)
		if out, err := formatCmd.CombinedOutput(); err != nil {
			return "", fmt.Errorf("mkfs.ext4 failed: %v %s", err, string(out))
		}
	}

	//3. Create persistent mount point
	// Naming: /mnt/arcanas-disk-{devname} (e.g. sdb)
	devName := strings.TrimPrefix(device, "/dev/")
	mountPath := "/mnt/arcanas-disk-" + devName

	if err := exec.Command("sudo", "mkdir", "-p", mountPath).Run(); err != nil {
		return "", fmt.Errorf("failed to make dir %s: %v", mountPath, err)
	}

	//4. Mount it
	if out, err := exec.Command("sudo", "mount", device, mountPath).CombinedOutput(); err != nil {
		return "", fmt.Errorf("mount failed: %v %s", err, string(out))
	}

	//5. Add to fstab (for disk persistence)
	// Check if entry already exists to avoid duplicates
	cmd = exec.Command("grep", "-q", mountPath, "/etc/fstab")
	err = cmd.Run()
	if err != nil {
		// Not found in fstab, add it
		// UUID is safer, but device path used for simplicity
		fstabEntry := fmt.Sprintf("%s %s ext4 defaults 0 0\n", device, mountPath)
		exec.Command("sudo", "sh", "-c", fmt.Sprintf("echo '%s' >> /etc/fstab", fstabEntry)).Run()
	}

	return mountPath, nil
}

func createBindMountPool(req models.StoragePoolCreateRequest) error {
	if len(req.Devices) != 1 {
		return fmt.Errorf("bind mount pools require exactly one device")
	}

	// Verify source path exists
	sourcePath := req.Devices[0]
	if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
		return fmt.Errorf("source path %s does not exist", sourcePath)
	}

	// Ensure data directory exists and has proper permissions
	cmd := exec.Command("sudo", "mkdir", "-p", "/srv")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to create data directory /srv: %v, output: %s", err, string(output))
	}

	// Set ownership of data directory to arcanas user if it exists
	cmd = exec.Command("sudo", "chown", "-R", "arcanas:arcanas", "/srv")
	if output, err := cmd.CombinedOutput(); err != nil {
		// Log warning but don't fail if arcanas user doesn't exist
		fmt.Printf("Warning: failed to set data directory ownership: %v, output: %s\n", err, string(output))
	}

	// Create mount point using sudo
	mountPoint := "/srv/" + req.Name
	cmd = exec.Command("sudo", "mkdir", "-p", mountPoint)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create mount point: %v", err)
	}

	// Create bind mount
	cmd = exec.Command("sudo", "mount", "--bind", sourcePath, mountPoint)
	if output, err := cmd.CombinedOutput(); err != nil {
		// Cleanup mount point on failure
		exec.Command("sudo", "rmdir", mountPoint).Run()
		return fmt.Errorf("failed to create bind mount: %v, output: %s", err, string(output))
	}

	// Set permissions for the mounted pool
	cmd = exec.Command("sudo", "chown", "-R", "nobody:nogroup", mountPoint)
	if output, err := cmd.CombinedOutput(); err != nil {
		fmt.Printf("Warning: failed to set pool ownership: %v, output: %s\n", err, string(output))
	}

	cmd = exec.Command("sudo", "chmod", "-R", "0777", mountPoint)
	if output, err := cmd.CombinedOutput(); err != nil {
		fmt.Printf("Warning: failed to set pool permissions: %v, output: %s\n", err, string(output))
	}

	// Add to fstab for persistence
	fstabEntry := fmt.Sprintf("%s %s none bind 0 0\n", sourcePath, mountPoint)
	cmd = exec.Command("sudo", "sh", "-c", fmt.Sprintf("echo '%s' >> /etc/fstab", fstabEntry))
	if err := cmd.Run(); err != nil {
		// Log warning but don't fail - fstab update is secondary
		fmt.Printf("Warning: failed to add to fstab: %v\n", err)
	}

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
	cmd := exec.Command("sudo", "mkdir", "-p", "/srv")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to create data directory /srv: %v, output: %s", err, string(output))
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
	mountPoint := "/srv/" + req.Name
	cmd = exec.Command("sudo", "mkdir", "-p", mountPoint)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create mount point: %v", err)
	}

	// Mount the logical volume
	cmd = exec.Command("sudo", "mount", lvPath, mountPoint)
	if err := cmd.Run(); err != nil {
		exec.Command("sudo", "rmdir", mountPoint).Run()
		return fmt.Errorf("failed to mount logical volume: %v", err)
	}

	// Set permissions for the mounted pool so Samba and other services can access it
	// Set ownership to allow Samba to write (nogroup is the default guest account)
	cmd = exec.Command("sudo", "chown", "-R", "nobody:nogroup", mountPoint)
	if output, err := cmd.CombinedOutput(); err != nil {
		// Log warning but don't fail - this might not be critical
		fmt.Printf("Warning: failed to set pool ownership: %v, output: %s\n", err, string(output))
	}

	// Set permissions to allow read/write access
	cmd = exec.Command("sudo", "chmod", "-R", "0777", mountPoint)
	if output, err := cmd.CombinedOutput(); err != nil {
		// Log warning but don't fail
		fmt.Printf("Warning: failed to set pool permissions: %v, output: %s\n", err, string(output))
	}

	// Add to fstab for persistence
	fstabEntry := fmt.Sprintf("%s %s ext4 defaults 0 0\n", lvPath, mountPoint)
	cmd = exec.Command("sudo", "sh", "-c", fmt.Sprintf("echo '%s' >> /etc/fstab", fstabEntry))
	if err := cmd.Run(); err != nil {
		// Log warning but don't fail - fstab update is secondary
		fmt.Printf("Warning: failed to add to fstab: %v\n", err)
	}

	return nil
}

func UpdateStoragePool(name string, req models.StoragePoolCreateRequest) error {
	// For mergerfs, we need to remount with new config
	mountPoint := "/srv/" + name

	// Unmount first
	cmd := exec.Command("sudo", "umount", mountPoint)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to unmount pool: %v", err)
	}

	// Remount with new config
	devicesStr := strings.Join(req.Devices, ":")
	config := req.Config
	if config == "" {
		config = "defaults,allow_other,use_ino"
	}

	cmd = exec.Command("sudo", "mergerfs", devicesStr, mountPoint, "-o", config)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to remount mergerfs pool: %v", err)
	}

	// Update fstab entry
	// First remove old entry
	cmd = exec.Command("sudo", "sed", "-i", fmt.Sprintf("\\|%s|d", mountPoint), "/etc/fstab")
	if err := cmd.Run(); err != nil {
		// Log warning but don't fail - entry might not exist
		fmt.Printf("Warning: failed to remove from fstab: %v\n", err)
	}

	// Add new entry
	fstabEntry := fmt.Sprintf("%s %s fuse.mergerfs %s 0 0\n", devicesStr, mountPoint, config)
	cmd = exec.Command("sudo", "sh", "-c", fmt.Sprintf("echo '%s' >> /etc/fstab", fstabEntry))
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to update fstab: %v", err)
	}

	return nil
}

func DeleteStoragePool(name string) error {
	mountPoint := "/srv/" + name

	// Check if samba is running and stop it temporarily
	sambaRunning := false
	if cmd := exec.Command("sudo", "systemctl", "is-active", "smbd").Run(); cmd == nil {
		sambaRunning = true
		fmt.Printf("Stopping Samba for pool deletion...\n")
		if output, err := exec.Command("sudo", "systemctl", "stop", "smbd").CombinedOutput(); err != nil {
			return fmt.Errorf("failed to stop Samba: %v, output: %s", err, string(output))
		}
	}

	// Check if it's currently mounted and unmount if needed
	if isMounted(mountPoint) {
		// Try normal unmount first
		cmd := exec.Command("sudo", "umount", mountPoint)
		output, err := cmd.CombinedOutput()
		if err != nil {
			// If normal unmount fails, try lazy unmount
			fmt.Printf("Normal unmount failed, trying lazy unmount...\n")
			cmd = exec.Command("sudo", "umount", "-l", mountPoint)
			output, err = cmd.CombinedOutput()
			if err != nil {
				return fmt.Errorf("failed to unmount %s: %v, output: %s", mountPoint, err, string(output))
			}
		}
		fmt.Printf("Successfully unmounted %s\n", mountPoint)
	}

	// Remove from fstab if exists
	cmd := exec.Command("sudo", "sed", "-i", fmt.Sprintf("\\|%s|d", mountPoint), "/etc/fstab")
	if err := cmd.Run(); err != nil {
		// Log warning but don't fail - fstab entry might not exist
		fmt.Printf("Warning: failed to remove from fstab: %v\n", err)
	}

	// Remove pool directory and all contents
	cmd = exec.Command("sudo", "rm", "-rf", mountPoint)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to remove pool directory %s: %v, output: %s", mountPoint, err, string(output))
	}

	// Restart Samba if it was running
	if sambaRunning {
		fmt.Printf("Restarting Samba...\n")
		if output, err := exec.Command("sudo", "systemctl", "start", "smbd").CombinedOutput(); err != nil {
			return fmt.Errorf("failed to restart Samba: %v, output: %s", err, string(output))
		}
	}

	fmt.Printf("Successfully deleted pool %s\n", name)
	return nil
}

func isMounted(mountPoint string) bool {
	cmd := exec.Command("findmnt", "-n", mountPoint)
	err := cmd.Run()
	return err == nil
}

// CleanupLegacyPool removes a pool from the old /var/lib/arcanas/ location
// This is useful for migrating pools to the new /srv/ location
// TODO: Remove this function after migration period (v1.0.0 or later)
// DEPRECATED: This is temporary migration helper code
func CleanupLegacyPool(poolName string) error {
	legacyMountPoint := "/var/lib/arcanas/" + poolName

	// Check if it's mounted
	if isMounted(legacyMountPoint) {
		cmd := exec.Command("sudo", "umount", legacyMountPoint)
		if output, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("failed to unmount legacy pool: %v, output: %s", legacyMountPoint, err, string(output))
		}
	}

	// Remove from fstab if exists
	cmd := exec.Command("sudo", "sed", "-i", fmt.Sprintf("\\|%s|d", legacyMountPoint), "/etc/fstab")
	if err := cmd.Run(); err != nil {
		// Log warning but don't fail
		fmt.Printf("Warning: failed to remove legacy pool from fstab: %v\n", err)
	}

	// Remove the pool directory if it exists
	if _, err := os.Stat(legacyMountPoint); err == nil {
		cmd = exec.Command("sudo", "rm", "-rf", legacyMountPoint)
		if output, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("failed to remove legacy pool directory: %v, output: %s", legacyMountPoint, err, string(output))
		}
	}

	return nil
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
	var used int64

	// Try to get used space using du -s -b
	// This works reliably for both regular and FUSE filesystems
	duCmd := exec.Command("du", "-s", "-b", mountPoint)
	duOutput, err := duCmd.Output()
	if err == nil {
		lines := strings.Split(string(duOutput), "\n")
		if len(lines) > 0 {
			fields := strings.Fields(lines[0])
			if len(fields) > 0 {
				used, _ = strconv.ParseInt(fields[0], 10, 64)
			}
		}
	}

	// Try df first for total size (works for regular filesystems)
	cmd := exec.Command("df", "-B", "1", "--output=size,used", mountPoint)
	output, err := cmd.Output()
	if err == nil {
		lines := strings.Split(string(output), "\n")
		if len(lines) >= 2 {
			fields := strings.Fields(lines[1])
			if len(fields) >= 1 {
				size, _ := strconv.ParseInt(fields[0], 10, 64)
				duUsed, _ := strconv.ParseInt(fields[1], 10, 64)
				if size > 0 {
					if duUsed > used {
						used = duUsed
					}
					return size, used
				}
			}
		}
	}

	// For FUSE filesystems where df returns 0 or unreliable values,
	// we can't reliably get total capacity. Return used for both
	// or 0, used if we couldn't get any values
	return used, used
}
